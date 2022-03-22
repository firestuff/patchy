package storebus

import "context"
import "fmt"
import "io"
import "net"
import "net/http"
import "os"
import "testing"

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

    if created.Text != "foo" {
      t.Fatalf("%s", created.Text)
    }

    if created.Id == "" {
      t.Fatalf("empty Id")
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

    if created.Text != "foo" {
      t.Fatalf("%s", created.Text)
    }

    if created.Id == "" {
      t.Fatalf("empty Id")
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

    // Initial event (saved object)
    buf, err := readString(body)

    if buf != fmt.Sprintf(`event: testtype
data: {"id":"%s","text":"foo"}

`, created.Id) {
      t.Fatalf("%s", buf)
    }

    // Heartbeat (after 5 seconds)
    buf, err = readString(body)

    if buf != `event: heartbeat
data: {}

` {
      t.Fatalf("%s", buf)
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

    buf, err = readString(body)

    if buf != fmt.Sprintf(`event: testtype
data: {"id":"%s","text":"bar"}

`, created.Id) {
      t.Fatalf("%s", buf)
    }

    body.Close()
  })
}

func withAPI(t *testing.T, cb func(*testing.T, *API, string, *resty.Client)) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	api, err := NewAPI(dir, &APIConfig{
		Factory:   factory,
		Update:    update,
		MayCreate: mayCreate,
		MayUpdate: mayUpdate,
		MayRead:   mayRead,
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

func readString(r io.Reader) (string, error) {
	buf := make([]byte, 256)

	n, err := r.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf[:n]), nil
}

type TestType struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

func (tt *TestType) GetType() string {
	return "testtype"
}

func (tt *TestType) GetId() string {
	return tt.Id
}

func (tt *TestType) SetId(id string) {
	tt.Id = id
}

func factory(t string) (Object, error) {
	switch t {
	case "testtype":
		return &TestType{}, nil
	default:
		return nil, fmt.Errorf("Unsupported type: %s", t)
	}
}

func update(obj Object, patch Object) error {
	switch o := obj.(type) {

	case *TestType:
		p := patch.(*TestType)

		if p.Text != "" {
			o.Text = p.Text
		}

		return nil

	default:
		return fmt.Errorf("Unsupported type: %s", obj.GetType())

	}
}

func mayCreate(obj Object, r *http.Request) error {
	return nil
}

func mayUpdate(obj Object, patch Object, r *http.Request) error {
	return nil
}

func mayRead(obj Object, r *http.Request) error {
	return nil
}
