package api

import "bufio"
import "fmt"
import "testing"
import "time"

import "github.com/go-resty/resty/v2"
import "github.com/stretchr/testify/require"

func TestStream(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		created := &testType{}

		resp, err := c.R().
			SetBody(&testType{
				Text: "foo",
			}).
			SetResult(created).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		resp, err = c.R().
			SetDoNotParseResponse(true).
			SetHeader("Accept", "text/event-stream").
			Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		body := resp.RawBody()
		defer body.Close()

		scan := bufio.NewScanner(body)

		initial := &testType{}
		eventType, err := readEvent(scan, initial)
		require.Nil(t, err)
		require.Equal(t, "update", eventType)
		require.Equal(t, "foo", initial.Text)

		// Heartbeat (after 5 seconds)
		eventType, err = readEvent(scan, nil)
		require.Equal(t, "heartbeat", eventType)

		updated := &testType{}

		// Round trip PATCH -> SSE
		resp, err = c.R().
			SetBody(&testType{
				Text: "bar",
			}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		eventType, err = readEvent(scan, updated)
		require.Equal(t, "update", eventType)
		require.Equal(t, "bar", updated.Text)
	})
}

func TestStreamRace(t *testing.T) {
	t.Parallel()

	// Check that Subscribe always gets its first and second events in order
	// and without gaps

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		created := &testType{}

		resp, err := c.R().
			SetBody(&testType{
				Num: 1,
			}).
			SetResult(created).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())

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

					resp, err := c.R().
						SetBody(created).
						Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
					if err != nil {
						doneUpdate <- err
						return
					}
					if resp.IsError() {
						doneUpdate <- fmt.Errorf("%s", resp.Error())
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
					if resp.IsError() {
						doneStream <- fmt.Errorf("%s", resp.Error())
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
		require.Nil(t, err)

		close(quitUpdate)

		err = <-doneUpdate
		require.Nil(t, err)
	})
}
