package patchy_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/firestuff/patchy"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func withAPI(t *testing.T, cb func(*testing.T, *patchy.API, string, *resty.Client)) {
	// TODO: Add goroutine leak detection: https://github.com/uber-go/goleak
	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	defer os.RemoveAll(dir)

	api, err := patchy.NewFileStoreAPI(dir)
	require.Nil(t, err)

	patchy.Register[testType](api)

	mux := http.NewServeMux()
	// Test that prefix stripping works
	mux.Handle("/api/", http.StripPrefix("/api", api))

	listener, err := net.Listen("tcp", "[::]:0")
	require.Nil(t, err)

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 1 * time.Second,
	}

	go func() {
		_ = srv.Serve(listener)
	}()

	baseURL := fmt.Sprintf("http://[::1]:%d/api", listener.Addr().(*net.TCPAddr).Port)

	c := resty.New().
		SetHeader("Content-Type", "application/json")

	cb(t, api, baseURL, c)

	err = srv.Shutdown(context.Background())
	require.Nil(t, err)
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

type testType struct {
	patchy.Metadata
	Text string `json:"text"`
	Num  int64  `json:"num"`
}

func TestCheckSafe(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	defer os.RemoveAll(dir)

	api, err := patchy.NewFileStoreAPI(dir)
	require.Nil(t, err)

	require.Nil(t, api.IsSafe())

	patchy.Register[testType](api)

	require.ErrorIs(t, api.IsSafe(), patchy.ErrMissingAuthCheck)

	require.Panics(t, func() {
		api.CheckSafe()
	})
}
