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

type testAPI struct {
	dir string
	api *patchy.API
	srv *http.Server
	rst *resty.Client
}

func newTestAPI(t *testing.T) *testAPI {
	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)

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

	baseURL := fmt.Sprintf("http://[::1]:%d/api/", listener.Addr().(*net.TCPAddr).Port)

	rst := resty.New().
		SetHeader("Content-Type", "application/json").
		SetBaseURL(baseURL)

	return &testAPI{
		dir: dir,
		api: api,
		srv: srv,
		rst: rst,
	}
}

func (ta *testAPI) r() *resty.Request {
	return ta.rst.R()
}

func (ta *testAPI) shutdown(t *testing.T) {
	err := ta.srv.Shutdown(context.Background())
	require.Nil(t, err)

	os.RemoveAll(ta.dir)
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

func TestAccept(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	created := &testType{}

	resp, err := ta.r().
		SetBody(&testType{
			Text: "foo",
		}).
		SetResult(created).
		Post("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "foo", created.Text)
	require.NotEmpty(t, created.ID)

	read := &testType{}

	resp, err = ta.r().
		SetHeader("Accept", "text/event-stream;q=0.3, text/xml;q=0.1, application/json;q=0.5").
		SetResult(read).
		SetPathParam("id", created.ID).
		Get("testtype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "application/json", resp.Header().Get("Content-Type"))
	require.Equal(t, "foo", read.Text)
	require.Equal(t, created.ID, read.ID)

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream;q=0.7, text/xml;q=0.1, application/json;q=0.5").
		SetPathParam("id", created.ID).
		Get("testtype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "text/event-stream", resp.Header().Get("Content-Type"))
	resp.RawBody().Close()
}
