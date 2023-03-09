package potency_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/firestuff/patchy/potency"
	"github.com/firestuff/patchy/store"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGET(t *testing.T) {
	t.Parallel()

	withServer(t, func(t *testing.T, url string, c *resty.Client) {
		key1 := uuid.NewString()

		resp, err := c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			Get(url)
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "bar", resp.Header().Get("X-Response"))

		resp1 := resp.String()

		resp, err = c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			Get(url)
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "bar", resp.Header().Get("X-Response"))
		require.Equal(t, resp1, resp.String())

		key2 := uuid.NewString()

		resp, err = c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key2)).
			Get(url)
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "bar", resp.Header().Get("X-Response"))

		resp2 := resp.String()

		require.NotEqual(t, resp2, resp1)

		resp, err = c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			Get(fmt.Sprintf("%sx", url))
		require.Nil(t, err)
		require.True(t, resp.IsError())

		resp, err = c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			Delete(url)
		require.Nil(t, err)
		require.True(t, resp.IsError())

		resp, err = c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			SetHeader("Authorization", "Bearer xyz").
			Get(url)
		require.Nil(t, err)
		require.True(t, resp.IsError())

		resp, err = c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			SetHeader("Accept", "text/xml").
			Get(url)
		require.Nil(t, err)
		require.True(t, resp.IsError())

		resp, err = c.R().
			SetHeader("Idempotency-Key", fmt.Sprintf(`"%s"`, key1)).
			SetHeader("X-Test", "foo").
			Get(url)
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "bar", resp.Header().Get("X-Response"))
		require.Equal(t, resp1, resp.String())
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
	// TODO: Switch this from callback to struct/defer Close()
	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	defer os.RemoveAll(dir)

	store := store.NewFileStore(dir)
	mux := http.NewServeMux()
	p := potency.NewPotency(store, mux)

	listener, err := net.Listen("tcp", "[::]:0")
	require.Nil(t, err)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		require.Nil(t, err)

		w.Header().Add("X-Response", "bar")

		_, err = w.Write([]byte(uuid.NewString()))
		require.Nil(t, err)
	})

	srv := &http.Server{
		Handler:           p,
		ReadHeaderTimeout: 1 * time.Second,
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
