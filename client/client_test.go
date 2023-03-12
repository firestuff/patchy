package client_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/firestuff/patchy/api"
	"github.com/firestuff/patchy/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type testType struct {
	api.Metadata
	Text *string
}

func TestClient(t *testing.T) {
	ctx := context.Background()

	dbname := fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.NewString())

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

	c := client.NewClient(baseURL)

	create, err := client.Create[testType](ctx, c, &testType{Text: client.P("foo")})
	require.Nil(t, err)
	require.Equal(t, "foo", *create.Text)

	get, err := client.Get[testType](ctx, c, create.ID)
	require.Nil(t, err)
	require.Equal(t, "foo", *get.Text)

	update, err := client.Update[testType](ctx, c, create.ID, &testType{Text: client.P("bar")})
	require.Nil(t, err)
	require.Equal(t, "bar", *update.Text)

	list, err := client.List[testType](ctx, c, nil)
	require.Nil(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "bar", *list[0].Text)

	find, err := client.Find[testType](ctx, c, create.ID[:4])
	require.Nil(t, err)
	require.Equal(t, "bar", *find.Text)
}
