package patchyc_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/dchest/uniuri"
	"github.com/firestuff/patchy/api"
	"github.com/firestuff/patchy/patchyc"
	"github.com/stretchr/testify/require"
)

type testType struct {
	api.Metadata
	Text1 *string
	Text2 *string
}

func TestClient(t *testing.T) {
	ctx := context.Background()

	dbname := fmt.Sprintf("file:%s?mode=memory&cache=shared", uniuri.New())

	a, err := api.NewSQLiteAPI(dbname)
	require.Nil(t, err)
	defer a.Close()

	api.Register[testType](a)

	mux := http.NewServeMux()
	// Test that prefix stripping works
	mux.Handle("/api/", http.StripPrefix("/api", a))

	listener, err := net.Listen("tcp", "[::]:0")
	require.Nil(t, err)

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 1 * time.Second,
	}

	go func() {
		_ = srv.Serve(listener)
	}()

	defer srv.Shutdown(ctx)

	baseURL := fmt.Sprintf("http://[::1]:%d/api/", listener.Addr().(*net.TCPAddr).Port)

	c := patchyc.NewClient(baseURL)

	create, err := patchyc.Create(ctx, c, &testType{
		Text1: patchyc.P("foo"),
		Text2: patchyc.P("zig"),
	})
	require.Nil(t, err)
	require.Equal(t, "foo", *create.Text1)

	get, err := patchyc.Get[testType](ctx, c, create.ID)
	require.Nil(t, err)
	require.Equal(t, create.ID, get.ID)
	require.Equal(t, "foo", *get.Text1)

	update, err := patchyc.Update(ctx, c, create.ID, &testType{Text1: patchyc.P("bar")})
	require.Nil(t, err)
	require.Equal(t, create.ID, update.ID)
	require.Equal(t, "bar", *update.Text1)
	require.Equal(t, "zig", *update.Text2)

	list, err := patchyc.List[testType](ctx, c, nil)
	require.Nil(t, err)
	require.Len(t, list, 1)
	require.Equal(t, create.ID, list[0].ID)
	require.Equal(t, "bar", *list[0].Text1)

	replace, err := patchyc.Replace(ctx, c, create.ID, &testType{Text1: patchyc.P("baz")})
	require.Nil(t, err)
	require.Equal(t, create.ID, replace.ID)
	require.Equal(t, "baz", *replace.Text1)
	require.Nil(t, replace.Text2)

	find, err := patchyc.Find[testType](ctx, c, create.ID[:4])
	require.Nil(t, err)
	require.Equal(t, create.ID, find.ID)
	require.Equal(t, "baz", *find.Text1)
	require.Nil(t, find.Text2)

	err = patchyc.Delete[testType](ctx, c, create.ID)
	require.Nil(t, err)

	list, err = patchyc.List[testType](ctx, c, nil)
	require.Nil(t, err)
	require.Len(t, list, 0)
}
