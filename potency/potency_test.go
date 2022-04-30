package potency

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/firestuff/patchy/store"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestGET(t *testing.T) {
	t.Parallel()

	withServer(t, func(t *testing.T, url string, c *resty.Client) {
		key1 := uuid.NewString()

		resp1a, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			Get(url)
		require.Nil(t, err)
		require.False(t, resp1a.IsError())

		resp1b, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			Get(url)
		require.Nil(t, err)
		require.False(t, resp1b.IsError())
		require.Equal(t, resp1a.String(), resp1b.String())

		key2 := uuid.NewString()

		resp2, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key2)).
			Get(url)
		require.Nil(t, err)
		require.False(t, resp2.IsError())
		require.NotEqual(t, resp1a.String(), resp2.String())

		resp1c, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			Get(fmt.Sprintf("%sx", url))
		require.Nil(t, err)
		require.True(t, resp1c.IsError())

		resp1d, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			Delete(url)
		require.Nil(t, err)
		require.True(t, resp1d.IsError())
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
		require.Nil(t, err)
		require.False(t, resp1a.IsError())

		resp1b, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			SetBody("test1").
			Post(url)
		require.Nil(t, err)
		require.False(t, resp1b.IsError())
		require.Equal(t, resp1b.String(), resp1a.String())

		resp1c, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			SetBody("test2").
			Post(url)
		require.Nil(t, err)
		require.True(t, resp1c.IsError())
	})
}

func withServer(t *testing.T, cb func(*testing.T, string, *resty.Client)) {
	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	store := store.NewFileStore(dir)
	p := NewPotency(store)

	listener, err := net.Listen("tcp", "[::]:0")
	require.Nil(t, err)

	router := mux.NewRouter()
	router.Use(p.Middleware)
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := ioutil.ReadAll(r.Body)
		require.Nil(t, err)

		_, err = w.Write([]byte(uuid.NewString()))
		require.Nil(t, err)
	})

	srv := &http.Server{
		Handler: router,
	}

	go func() {
		_ = srv.Serve(listener)
	}()

	url := fmt.Sprintf("http://[::1]:%d/", listener.Addr().(*net.TCPAddr).Port)

	c := resty.New().
		SetHeader("Content-Type", "application/json")

	cb(t, url, c)

	err = srv.Shutdown(context.Background())
	require.Nil(t, err)
}
