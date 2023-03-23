package api_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/dchest/uniuri"
	"github.com/firestuff/patchy"
	"github.com/firestuff/patchy/api"
	"github.com/firestuff/patchy/patchyc"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

type testAPI struct {
	api *api.API
	srv *http.Server
	rst *resty.Client
	pyc *patchyc.Client
}

func newTestAPI(t *testing.T) *testAPI {
	dbname := fmt.Sprintf("file:%s?mode=memory&cache=shared", uniuri.New())

	a, err := api.NewSQLiteAPI(dbname)
	require.NoError(t, err)

	api.Register[testType](a)

	mux := http.NewServeMux()
	// Test that prefix stripping works
	mux.Handle("/api/", http.StripPrefix("/api", a))

	listener, err := net.Listen("tcp", "[::]:0")
	require.NoError(t, err)

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

	pyc := patchyc.NewClient(baseURL)

	return &testAPI{
		api: a,
		srv: srv,
		rst: rst,
		pyc: pyc,
	}
}

func (ta *testAPI) r() *resty.Request {
	return ta.rst.R()
}

func (ta *testAPI) shutdown(t *testing.T) {
	err := ta.srv.Shutdown(context.Background())
	require.NoError(t, err)

	ta.api.Close()
}

type testType struct {
	api.Metadata
	Text string `json:"text"`
	Num  int64  `json:"num"`
}

type testType2 struct {
	api.Metadata
	Text string `json:"text"`
}

type testType3 struct {
	api.Metadata
	Text string `json:"text"`
}

type missingMetadata struct {
	Text string `json:"text"`
}

func (tt *testType) MayRead(context.Context, *api.API) error {
	return nil
}

func (tt *testType2) MayWrite(context.Context, *testType2, *api.API) error {
	return nil
}

func (tt *testType3) MayRead(context.Context, *api.API) error {
	return nil
}

func (tt *testType3) MayWrite(context.Context, *testType3, *api.API) error {
	return nil
}

func TestFileStoreAPI(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	a, err := api.NewFileStoreAPI(dir)
	require.NoError(t, err)

	api.Register[testType](a)

	ctx := context.Background()

	created, err := api.Create(ctx, a, &testType{Text: "foo"})
	require.NoError(t, err)

	get, err := api.Get[testType](ctx, a, created.ID)
	require.NoError(t, err)
	require.NotNil(t, get)
	require.Equal(t, "foo", get.Text)
}

func TestRegisterMissingMetadata(t *testing.T) {
	t.Parallel()

	dbname := fmt.Sprintf("file:%s?mode=memory&cache=shared", uniuri.New())

	a, err := api.NewSQLiteAPI(dbname)
	require.NoError(t, err)

	defer a.Close()

	require.Panics(t, func() {
		api.Register[missingMetadata](a)
	})
}

func TestIsSafeSuccess(t *testing.T) {
	t.Parallel()

	dbname := fmt.Sprintf("file:%s?mode=memory&cache=shared", uniuri.New())

	a, err := api.NewSQLiteAPI(dbname)
	require.NoError(t, err)

	defer a.Close()

	api.Register[testType3](a)

	require.NoError(t, a.IsSafe())
}

func TestIsSafeWithoutWrite(t *testing.T) {
	t.Parallel()

	dbname := fmt.Sprintf("file:%s?mode=memory&cache=shared", uniuri.New())

	a, err := api.NewSQLiteAPI(dbname)
	require.NoError(t, err)

	defer a.Close()

	require.NoError(t, a.IsSafe())

	api.Register[testType](a)

	require.ErrorIs(t, a.IsSafe(), api.ErrMissingAuthCheck)
}

func TestIsSafeWithoutRead(t *testing.T) {
	t.Parallel()

	dbname := fmt.Sprintf("file:%s?mode=memory&cache=shared", uniuri.New())

	a, err := api.NewSQLiteAPI(dbname)
	require.NoError(t, err)

	defer a.Close()

	require.NoError(t, a.IsSafe())

	api.Register[testType2](a)

	require.ErrorIs(t, a.IsSafe(), api.ErrMissingAuthCheck)
}

func TestCheckSafeSuccess(t *testing.T) {
	t.Parallel()

	dbname := fmt.Sprintf("file:%s?mode=memory&cache=shared", uniuri.New())

	a, err := api.NewSQLiteAPI(dbname)
	require.NoError(t, err)

	defer a.Close()

	api.Register[testType3](a)

	require.NotPanics(t, a.CheckSafe)
}

func TestCheckSafeWithoutWrite(t *testing.T) {
	t.Parallel()

	dbname := fmt.Sprintf("file:%s?mode=memory&cache=shared", uniuri.New())

	a, err := api.NewSQLiteAPI(dbname)
	require.NoError(t, err)

	defer a.Close()

	require.NotPanics(t, a.CheckSafe)

	api.Register[testType](a)

	require.Panics(t, a.CheckSafe)
}

func TestCheckSafeWithoutRead(t *testing.T) {
	t.Parallel()

	dbname := fmt.Sprintf("file:%s?mode=memory&cache=shared", uniuri.New())

	a, err := api.NewSQLiteAPI(dbname)
	require.NoError(t, err)

	defer a.Close()

	require.NotPanics(t, a.CheckSafe)

	api.Register[testType2](a)

	require.Panics(t, a.CheckSafe)
}

func TestAcceptJSON(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	read := &testType{}

	resp, err := ta.r().
		SetHeader("Accept", "text/event-stream;q=0.3, text/xml;q=0.1, application/json;q=0.5").
		SetResult(read).
		SetPathParam("id", created.ID).
		Get("testtype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "application/json", resp.Header().Get("Content-Type"))
	require.Equal(t, "foo", read.Text)
	require.Equal(t, created.ID, read.ID)
}

func TestAcceptEventStream(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	resp, err := ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream;q=0.7, text/xml;q=0.1, application/json;q=0.5").
		SetPathParam("id", created.ID).
		Get("testtype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "text/event-stream", resp.Header().Get("Content-Type"))
	resp.RawBody().Close()
}

func TestAcceptFailure(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	resp, err := ta.r().
		SetHeader("Accept", "unsupported").
		SetPathParam("id", created.ID).
		Get("testtype/{id}")
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, http.StatusNotAcceptable, resp.StatusCode())
}

func TestAcceptListFailure(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	resp, err := ta.r().
		SetHeader("Accept", "unsupported").
		Get("testtype")
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, http.StatusNotAcceptable, resp.StatusCode())
}

func TestDebug(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	dbg := map[string]any{}

	resp, err := ta.r().
		SetResult(&dbg).
		Get("_debug")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Contains(t, dbg, "server")
	require.Contains(t, dbg, "ip")
	require.Contains(t, dbg, "http")
	require.Contains(t, dbg, "tls")
}

func TestRequestHookError(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	ta.api.SetRequestHook(func(*http.Request, *patchy.API) (*http.Request, error) {
		return nil, fmt.Errorf("test reject") //nolint:goerr113
	})

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.Error(t, err)
	require.Nil(t, created)
}
