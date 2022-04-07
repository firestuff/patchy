package api

import "bufio"
import "fmt"
import "testing"
import "time"

import "github.com/go-resty/resty/v2"

func TestGETStream(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		created := &testType{}

		_, err := c.R().
			SetBody(&testType{
				Text: "foo",
			}).
			SetResult(created).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		if err != nil {
			t.Fatal(err)
		}

		resp, err := c.R().
			SetDoNotParseResponse(true).
			SetHeader("Accept", "text/event-stream").
			Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}
		body := resp.RawBody()
		defer body.Close()

		scan := bufio.NewScanner(body)

		initial := &testType{}
		eventType, err := readEvent(scan, initial)
		if err != nil {
			t.Fatal(err)
		}

		if eventType != "update" {
			t.Error(eventType)
		}

		if initial.Text != "foo" {
			t.Error(initial)
		}

		// Heartbeat (after 5 seconds)
		eventType, err = readEvent(scan, nil)

		if eventType != "heartbeat" {
			t.Error(eventType)
		}

		updated := &testType{}

		// Round trip PATCH -> SSE
		_, err = c.R().
			SetBody(&testType{
				Text: "bar",
			}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}

		eventType, err = readEvent(scan, updated)

		if eventType != "update" {
			t.Error(eventType)
		}

		if updated.Text != "bar" {
			t.Error(eventType)
		}

		body.Close()
	})
}

func TestGETStreamRace(t *testing.T) {
	t.Parallel()

	// Check that Subscribe always gets its first and second events in order
	// and without gaps

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		created := &testType{}

		_, err := c.R().
			SetBody(&testType{
				Num: 1,
			}).
			SetResult(created).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		if err != nil {
			t.Fatal(err)
		}

		quitUpdate := make(chan bool)
		doneUpdate := make(chan error)
		quitStream := make(chan bool)
		doneStream := make(chan error)

		go func() {
			for {
				select {
				case <-quitUpdate:
					close(doneUpdate)
					return
				default:
					created.Num++

					_, err := c.R().
						SetBody(created).
						Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
					if err != nil {
						doneUpdate <- err
						return
					}
				}
			}
		}()

		go func() {
			for {
				select {
				case <-quitStream:
					close(doneStream)
					return
				default:
					resp, err := c.R().
						SetDoNotParseResponse(true).
						SetHeader("Accept", "text/event-stream").
						Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
					if err != nil {
						doneStream <- err
						return
					}
					body := resp.RawBody()
					defer body.Close()

					scan := bufio.NewScanner(body)

					first := &testType{}
					eventType, err := readEvent(scan, first)
					if err != nil {
						doneStream <- err
						return
					}

					if eventType != "update" {
						doneStream <- fmt.Errorf("update(1) != %s", eventType)
						return
					}

					second := &testType{}
					eventType, err = readEvent(scan, second)
					if err != nil {
						doneStream <- err
						return
					}

					if eventType != "update" {
						doneStream <- fmt.Errorf("update(2) != %s", eventType)
						return
					}

					if second.Num != first.Num+1 {
						doneStream <- fmt.Errorf("%+v %+v", first, second)
						return
					}

					body.Close()
				}
			}
		}()

		time.Sleep(3 * time.Second)

		close(quitStream)

		err = <-doneStream
		if err != nil {
			t.Fatal(err)
		}

		close(quitUpdate)

		err = <-doneUpdate
		if err != nil {
			t.Fatal(err)
		}
	})
}
