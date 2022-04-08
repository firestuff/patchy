package potency

import "context"
import "fmt"
import "io/ioutil"
import "net"
import "net/http"
import "os"
import "testing"

import "github.com/firestuff/patchy/store"
import "github.com/google/uuid"
import "github.com/gorilla/mux"
import "github.com/go-resty/resty/v2"

func TestGET(t *testing.T) {
	t.Parallel()

	withServer(t, func(t *testing.T, url string, c *resty.Client) {
		key1 := uuid.NewString()

		resp1a, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			Get(url)
		if err != nil {
			t.Fatal(err)
		}
		if resp1a.IsError() {
			t.Fatal(resp1a)
		}

		resp1b, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			Get(url)
		if err != nil {
			t.Fatal(err)
		}
		if resp1b.IsError() {
			t.Fatal(resp1b)
		}

		if resp1a.String() != resp1b.String() {
			t.Fatalf("%s vs %s", resp1a, resp1b)
		}

		key2 := uuid.NewString()

		resp2, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key2)).
			Get(url)
		if err != nil {
			t.Fatal(err)
		}
		if resp2.IsError() {
			t.Fatal(resp2)
		}

		if resp1a.String() == resp2.String() {
			t.Fatal(resp1a)
		}

		resp1c, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			Get(fmt.Sprintf("%sx", url))
		if err != nil {
			t.Fatal(err)
		}
		if !resp1c.IsError() {
			t.Fatal("Improper success")
		}

		resp1d, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			Delete(url)
		if err != nil {
			t.Fatal(err)
		}
		if !resp1d.IsError() {
			t.Fatal("Improper success")
		}
	})
}

func TestPOST(t *testing.T) {
	t.Parallel()

	withServer(t, func(t *testing.T, url string, c *resty.Client) {
		key1 := uuid.NewString()

		resp1a, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			SetBody("test1").
			Post(url)
		if err != nil {
			t.Fatal(err)
		}
		if resp1a.IsError() {
			t.Fatal(resp1a)
		}

		resp1b, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			SetBody("test1").
			Post(url)
		if err != nil {
			t.Fatal(err)
		}
		if resp1b.IsError() {
			t.Fatal(resp1b)
		}

		if resp1a.String() != resp1b.String() {
			t.Fatalf("%s vs %s", resp1a, resp1b)
		}

		resp1c, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			SetBody("test2").
			Post(url)
		if err != nil {
			t.Fatal(err)
		}
		if !resp1c.IsError() {
			t.Fatal(resp1c)
		}
	})
}

func withServer(t *testing.T, cb func(*testing.T, string, *resty.Client)) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store := store.NewLocalStore(dir)
	p := NewPotency(store)

	listener, err := net.Listen("tcp", "[::]:0")
	if err != nil {
		t.Fatal(err)
	}

	router := mux.NewRouter()
	router.Use(p.Middleware)
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ioutil.ReadAll(r.Body)
		w.Write([]byte(uuid.NewString()))
	})

	srv := &http.Server{
		Handler: router,
	}

	go srv.Serve(listener)

	url := fmt.Sprintf("http://[::1]:%d/", listener.Addr().(*net.TCPAddr).Port)

	c := resty.New().
		SetHeader("Content-Type", "application/json")

	cb(t, url, c)

	srv.Shutdown(context.Background())
}
