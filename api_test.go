package storebus

import "context"
import "fmt"
import "net"
import "net/http"
import "os"
import "testing"

import "github.com/go-resty/resty/v2"

func TestAPI(t *testing.T) {
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

	urlBase := fmt.Sprintf("http://[::1]:%d/api", listener.Addr().(*net.TCPAddr).Port)

	c := resty.New().
		SetHeader("Content-Type", "application/json")

	resp := &TestType{}

	_, err = c.R().
		SetBody(&TestType{
			Text: "foo",
		}).
		SetResult(resp).
		Post(fmt.Sprintf("%s/testtype", urlBase))
	if err != nil {
		t.Fatal(err)
	}

	if resp.Text != "foo" {
		t.Fatalf("%s", resp.Text)
	}

	if resp.Id == "" {
		t.Fatalf("empty Id")
	}

	srv.Shutdown(context.Background())
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
		return nil, fmt.Errorf("Unknown type: %s", t)
	}
}

func update(obj Object, newObj Object) error {
	return nil
}

func mayCreate(obj Object, r *http.Request) error {
	return nil
}

func mayUpdate(obj Object, newObj Object, r *http.Request) error {
	return nil
}

func mayRead(obj Object, r *http.Request) error {
	return nil
}
