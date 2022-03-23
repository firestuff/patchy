package storebus

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
		doneUpdate := make(chan bool)
		quitStream := make(chan bool)
		doneStream := make(chan bool)

		go func() {
			for {
				select {
				case <-quitUpdate:
					close(doneUpdate)
					return
				default:
					created.Num++

					_, err = c.R().
						SetBody(created).
						Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
					if err != nil {
						t.Fatal(err)
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
						t.Fatal(err)
					}
					body := resp.RawBody()
					defer body.Close()

					scan := bufio.NewScanner(body)

					first := &TestType{}
					eventType, err := readEvent(scan, first)
					if err != nil {
						t.Fatal(err)
					}

					if eventType != "update" {
						t.Fatal(eventType)
					}

					second := &TestType{}
					eventType, err = readEvent(scan, second)
					if err != nil {
						t.Fatal(err)
					}

					if eventType != "update" {
						t.Fatal(eventType)
					}

					if second.Num != first.Num+1 {
						t.Fatalf("%+v %+v", first, second)
					}

					body.Close()
				}
			}
		}()

		time.Sleep(3 * time.Second)
		close(quitStream)
		<-doneStream
		close(quitUpdate)
		<-doneUpdate
	})
}

func withAPI(t *testing.T, cb func(*testing.T, *API, string, *resty.Client)) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	api, err := NewAPI(dir, map[string]*APIConfig{
		"testtype": &APIConfig{
			Factory:   factory,
			Update:    update,
			MayCreate: mayCreate,
			MayUpdate: mayUpdate,
			MayDelete: mayDelete,
			MayRead:   mayRead,
		},
	})
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

func readEvent(scan *bufio.Scanner, out interface{}) (string, error) {
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
	Id   string `json:"id"`
	Text string `json:"text"`
	Num  int64  `json:"num"`
}

func (tt *TestType) GetId() string {
	return tt.Id
}

func (tt *TestType) SetId(id string) {
	tt.Id = id
}

func factory() (Object, error) {
	return &TestType{}, nil
}

func update(obj Object, patch Object) error {
	o := obj.(*TestType)
	p := patch.(*TestType)

	if p.Text != "" {
		o.Text = p.Text
	}

	if p.Num != 0 {
		o.Num = p.Num
	}

	return nil
}

func mayCreate(obj Object, r *http.Request) error {
	return nil
}

func mayUpdate(obj Object, patch Object, r *http.Request) error {
	return nil
}

func mayDelete(obj Object, r *http.Request) error {
	return nil
}

func mayRead(obj Object, r *http.Request) error {
	return nil
}
