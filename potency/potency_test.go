package potency

import "context"
import "fmt"
import "net"
import "net/http"
import "os"
import "testing"

import "github.com/firestuff/patchy/store"
import "github.com/google/uuid"
import "github.com/gorilla/mux"
import "github.com/go-resty/resty/v2"

func TestGet(t *testing.T) {
	t.Parallel()

	withServer(t, func(t *testing.T, url string, c *resty.Client) {
		key1 := uuid.NewString()

		resp1a, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			Get(url)
		if err != nil {
			t.Fatal(err)
		}

		resp1b, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			Get(url)
		if err != nil {
			t.Fatal(err)
		}

		if resp1a.String() != resp1b.String() {
			t.Fatalf("%s vs %s", resp1a, resp1b)
		}
	})
}

func withServer(t *testing.T, cb func(*testing.T, string, *resty.Client)) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store := store.NewStore(dir)
	p := NewPotency(store)

	listener, err := net.Listen("tcp", "[::]:0")
	if err != nil {
		t.Fatal(err)
	}

	router := mux.NewRouter()
	router.Use(p.Middleware)
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
