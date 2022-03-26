package api

import "bufio"
import "bytes"
import "context"
import "encoding/json"
import "fmt"
import "io"
import "net"
import "net/http"
import "os"
import "strings"
import "testing"
import "time"

import "github.com/go-resty/resty/v2"

import "github.com/firestuff/patchy/metadata"

func TestAPICreate(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		created := &TestType{}

		_, err := c.R().
			SetBody(&TestType{
				Text: "foo",
			}).
			SetResult(created).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		if err != nil {
			t.Fatal(err)
		}

		if created.Text != "foo" {
			t.Fatalf("%s", created.Text)
		}

		if created.Id == "" {
			t.Fatalf("empty Id")
		}

		read := &TestType{}

		_, err = c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}

		if read.Text != "foo" {
			t.Fatalf("%s", read.Text)
		}

		if read.Id != read.Id {
			t.Fatalf("%s %s", read.Id, created.Id)
		}
	})
}

func TestAPIUpdate(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		created := &TestType{}

		_, err := c.R().
			SetBody(&TestType{
				Text: "foo",
			}).
			SetResult(created).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		if err != nil {
			t.Fatal(err)
		}

		updated := &TestType{}

		_, err = c.R().
			SetBody(&TestType{
				Text: "bar",
			}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}

		if updated.Text != "bar" {
			t.Fatalf("%s", updated.Text)
		}

		if updated.Id != created.Id {
			t.Fatalf("%s %s", updated.Id, created.Id)
		}

		read := &TestType{}

		_, err = c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}

		if read.Text != "bar" {
			t.Fatalf("%s", read.Text)
		}

		if read.Id != read.Id {
			t.Fatalf("%s %s", read.Id, created.Id)
		}
	})
}

func TestAPIDelete(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		created := &TestType{}

		_, err := c.R().
			SetBody(&TestType{
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

		initial := &TestType{}
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

		_, err = c.R().
			Delete(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}

		eventType, err = readEvent(scan, nil)

		if eventType != "delete" {
			t.Error(eventType)
		}

		body.Close()

		read := &TestType{}

		resp, err = c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if !resp.IsError() {
			t.Fatal(read)
		}
	})
}

func TestAPIStream(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		created := &TestType{}

		_, err := c.R().
			SetBody(&TestType{
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

		initial := &TestType{}
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

		updated := &TestType{}

		// Round trip PATCH -> SSE
		_, err = c.R().
			SetBody(&TestType{
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

func TestAPIStreamRace(t *testing.T) {
	t.Parallel()

	// Check that Subscribe always gets its first and second events in order
	// and without gaps

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		created := &TestType{}

		_, err := c.R().
			SetBody(&TestType{
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

					first := &TestType{}
					eventType, err := readEvent(scan, first)
					if err != nil {
						doneStream <- err
						return
					}

					if eventType != "update" {
						doneStream <- fmt.Errorf("update(1) != %s", eventType)
						return
					}

					second := &TestType{}
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

func withAPI(t *testing.T, cb func(*testing.T, *API, string, *resty.Client)) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	api, err := NewAPI(dir)
	if err != nil {
		t.Fatal(err)
	}

	err = Register[TestType](api, "testtype", func() *TestType { return &TestType{} })
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	// Test that prefix stripping works
	mux.Handle("/api/", http.StripPrefix("/api", api))

	listener, err := net.Listen("tcp", "[::]:0")
	if err != nil {
		t.Fatal(err)
	}

	srv := &http.Server{
		Handler: mux,
	}

	go srv.Serve(listener)

	baseURL := fmt.Sprintf("http://[::1]:%d/api", listener.Addr().(*net.TCPAddr).Port)

	c := resty.New().
		SetHeader("Content-Type", "application/json")

	cb(t, api, baseURL, c)

	srv.Shutdown(context.Background())
}

func readEvent(scan *bufio.Scanner, out any) (string, error) {
	eventType := ""
	data := [][]byte{}

	for scan.Scan() {
		line := scan.Text()

		switch {

		case strings.HasPrefix(line, ":"):
			continue

		case strings.HasPrefix(line, "event: "):
			eventType = strings.TrimPrefix(line, "event: ")

		case strings.HasPrefix(line, "data: "):
			data = append(data, bytes.TrimPrefix(scan.Bytes(), []byte("data: ")))

		case line == "":
			var err error

			if out != nil {
				err = json.Unmarshal(bytes.Join(data, []byte("\n")), out)
			}

			return eventType, err

		}
	}

	return "", io.EOF
}

type TestType struct {
	metadata.Metadata
	Text string `json:"text"`
	Num  int64  `json:"num"`
}

func (tt *TestType) MayCreate(r *http.Request) error {
	return nil
}

func (tt *TestType) MayUpdate(patch *TestType, r *http.Request) error {
	return nil
}

func (tt *TestType) MayDelete(r *http.Request) error {
	return nil
}

func (tt *TestType) MayRead(r *http.Request) error {
	return nil
}
