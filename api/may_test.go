//nolint:goerr113
package api_test

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/firestuff/patchy/api"
	"github.com/stretchr/testify/require"
)

type mayType struct {
	api.Metadata
	Text1 string
}

func (mt *mayType) MayRead(ctx context.Context, a *api.API) error {
	if ctx.Value(refuseRead) != nil {
		return fmt.Errorf("may not read")
	}

	t1r := ctx.Value(text1Read)
	if t1r != nil {
		mt.Text1 = t1r.(string)
	}

	nt1 := ctx.Value(newText1)
	if nt1 != nil {
		// Use a separate context so we don't recursively create objects
		_, err := api.Create(context.Background(), a, &mayType{Text1: nt1.(string)}) //nolint:contextcheck
		if err != nil {
			return err
		}
	}

	return nil
}

func (mt *mayType) MayWrite(ctx context.Context, prev *mayType, a *api.API) error {
	if ctx.Value(refuseWrite) != nil {
		return fmt.Errorf("may not write")
	}

	t1w := ctx.Value(text1Write)
	if t1w != nil {
		mt.Text1 = t1w.(string)
	}

	return nil
}

type contextKey int

const (
	refuseRead contextKey = iota
	refuseWrite
	text1Read
	text1Write
	newText1
)

func requestHook(r *http.Request, _ *api.API) (*http.Request, error) {
	ctx := r.Context()

	if r.Header.Get("X-Refuse-Read") != "" {
		ctx = context.WithValue(ctx, refuseRead, true)
	}

	if r.Header.Get("X-Refuse-Write") != "" {
		ctx = context.WithValue(ctx, refuseWrite, true)
	}

	t1r := r.Header.Get("X-Text1-Read")
	if t1r != "" {
		ctx = context.WithValue(ctx, text1Read, t1r)
	}

	t1w := r.Header.Get("X-Text1-Write")
	if t1w != "" {
		ctx = context.WithValue(ctx, text1Write, t1w)
	}

	nt1 := r.Header.Get("X-NewText1")
	if nt1 != "" {
		ctx = context.WithValue(ctx, newText1, nt1)
	}

	return r.WithContext(ctx), nil
}

func TestMayWrite(t *testing.T) {
	// TODO: Break up test
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ta.api.SetRequestHook(requestHook)
	api.Register[mayType](ta.api)

	created := &mayType{}

	resp, err := ta.r().
		SetBody(&mayType{}).
		SetResult(created).
		Post("maytype")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.NotEmpty(t, created.ID)

	resp, err = ta.r().
		SetHeader("X-Refuse-Write", "x").
		SetBody(&mayType{}).
		SetResult(created).
		Post("maytype")
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, 401, resp.StatusCode())

	replaced := &mayType{}

	resp, err = ta.r().
		SetBody(&mayType{}).
		SetResult(replaced).
		SetPathParam("id", created.ID).
		Put("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	resp, err = ta.r().
		SetHeader("X-Refuse-Write", "x").
		SetBody(&mayType{}).
		SetResult(replaced).
		SetPathParam("id", created.ID).
		Put("maytype/{id}")
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, 401, resp.StatusCode())

	updated := &mayType{}

	resp, err = ta.r().
		SetBody(&mayType{}).
		SetResult(updated).
		SetPathParam("id", created.ID).
		Patch("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError(), resp)

	resp, err = ta.r().
		SetHeader("X-Refuse-Write", "x").
		SetBody(&mayType{}).
		SetResult(updated).
		SetPathParam("id", created.ID).
		Patch("maytype/{id}")
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, 401, resp.StatusCode())

	resp, err = ta.r().
		SetHeader("X-Refuse-Write", "x").
		SetPathParam("id", created.ID).
		Delete("maytype/{id}")
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, 401, resp.StatusCode())

	resp, err = ta.r().
		SetPathParam("id", created.ID).
		Delete("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
}

func TestMayRead(t *testing.T) {
	// TODO: Break up test
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ta.api.SetRequestHook(requestHook)
	api.Register[mayType](ta.api)

	created := &mayType{}

	resp, err := ta.r().
		SetBody(&mayType{}).
		SetResult(created).
		Post("maytype")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	read := &testType{}

	resp, err = ta.r().
		SetResult(read).
		SetPathParam("id", created.ID).
		Get("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetPathParam("id", created.ID).
		Get("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	resp.RawBody().Close()

	resp, err = ta.r().
		SetHeader("X-Refuse-Read", "x").
		SetResult(read).
		SetPathParam("id", created.ID).
		Get("maytype/{id}")
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, 401, resp.StatusCode())

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("X-Refuse-Read", "x").
		SetHeader("Accept", "text/event-stream").
		SetPathParam("id", created.ID).
		Get("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	resp.RawBody().Close()

	list := []*testType{}

	resp, err = ta.r().
		SetResult(&list).
		Get("maytype")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 1)

	resp, err = ta.r().
		SetHeader("X-Refuse-Read", "x").
		SetResult(&list).
		Get("maytype")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 0)

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		Get("maytype")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	body1 := resp.RawBody()
	defer body1.Close()

	scan1 := bufio.NewScanner(body1)

	eventType, err := readEvent(scan1, &list)
	require.NoError(t, err)
	require.Equal(t, "list", eventType)
	require.Len(t, list, 1)

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("X-Refuse-Read", "x").
		SetHeader("Accept", "text/event-stream").
		Get("maytype")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	body2 := resp.RawBody()
	defer body2.Close()

	scan2 := bufio.NewScanner(body2)

	eventType, err = readEvent(scan2, &list)
	require.NoError(t, err)
	require.Equal(t, "list", eventType)
	require.Len(t, list, 0)
}

func TestMayWriteMutate(t *testing.T) {
	// TODO: Break up test
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ta.api.SetRequestHook(requestHook)
	api.Register[mayType](ta.api)

	create := &mayType{}

	resp, err := ta.r().
		SetHeader("X-Text1-Write", "1234").
		SetBody(&mayType{Text1: "foo"}).
		SetResult(create).
		Post("maytype")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "1234", create.Text1)

	get := &mayType{}

	resp, err = ta.r().
		SetResult(get).
		SetPathParam("id", create.ID).
		Get("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "1234", get.Text1)

	patch := &mayType{}

	resp, err = ta.r().
		SetHeader("X-Text1-Write", "2345").
		SetBody(&mayType{Text1: "bar"}).
		SetResult(patch).
		SetPathParam("id", create.ID).
		Patch("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "2345", patch.Text1)

	resp, err = ta.r().
		SetResult(get).
		SetPathParam("id", create.ID).
		Get("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "2345", get.Text1)

	put := &mayType{}

	resp, err = ta.r().
		SetHeader("X-Text1-Write", "3456").
		SetBody(&mayType{Text1: "zig"}).
		SetResult(put).
		SetPathParam("id", create.ID).
		Put("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "3456", put.Text1)

	resp, err = ta.r().
		SetResult(get).
		SetPathParam("id", create.ID).
		Get("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "3456", get.Text1)
}

func TestMayReadMutate(t *testing.T) {
	// TODO: Break up test
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ta.api.SetRequestHook(requestHook)
	api.Register[mayType](ta.api)

	create := &mayType{}

	resp, err := ta.r().
		SetHeader("X-Text1-Read", "1234").
		SetBody(&mayType{Text1: "foo"}).
		SetResult(create).
		Post("maytype")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "1234", create.Text1)

	get := &mayType{}

	resp, err = ta.r().
		SetResult(get).
		SetPathParam("id", create.ID).
		Get("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "foo", get.Text1)

	patch := &mayType{}

	resp, err = ta.r().
		SetHeader("X-Text1-Read", "2345").
		SetBody(&mayType{Text1: "bar"}).
		SetResult(patch).
		SetPathParam("id", create.ID).
		Patch("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "2345", patch.Text1)

	resp, err = ta.r().
		SetResult(get).
		SetPathParam("id", create.ID).
		Get("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "bar", get.Text1)

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("X-Text1-Read", "stream1234").
		SetHeader("Accept", "text/event-stream").
		SetPathParam("id", create.ID).
		Get("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	body := resp.RawBody()
	defer body.Close()

	scan := bufio.NewScanner(body)

	stream := &mayType{}

	eventType, err := readEvent(scan, stream)
	require.NoError(t, err)
	require.Equal(t, "initial", eventType)
	require.Equal(t, "stream1234", stream.Text1)

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("X-Text1-Read", "stream2345").
		SetHeader("Accept", "text/event-stream").
		Get("maytype")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	bodyList := resp.RawBody()
	defer bodyList.Close()

	scanList := bufio.NewScanner(bodyList)

	streamList := []*mayType{}

	eventType, err = readEvent(scanList, &streamList)
	require.NoError(t, err)
	require.Equal(t, "list", eventType)
	require.Len(t, streamList, 1)
	require.Equal(t, "stream2345", streamList[0].Text1)

	put := &mayType{}

	resp, err = ta.r().
		SetHeader("X-Text1-Read", "3456").
		SetBody(&mayType{Text1: "zig"}).
		SetResult(put).
		SetPathParam("id", create.ID).
		Put("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "3456", put.Text1)

	resp, err = ta.r().
		SetResult(get).
		SetPathParam("id", create.ID).
		Get("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "zig", get.Text1)

	resp, err = ta.r().
		SetHeader("X-Text1-Read", "4567").
		SetResult(get).
		SetPathParam("id", create.ID).
		Get("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "4567", get.Text1)

	eventType, err = readEvent(scan, stream)
	require.NoError(t, err)
	require.Equal(t, "update", eventType)
	require.Equal(t, "stream1234", stream.Text1)

	eventType, err = readEvent(scanList, &streamList)
	require.NoError(t, err)
	require.Equal(t, "list", eventType)
	require.Len(t, streamList, 1)
	require.Equal(t, "stream2345", streamList[0].Text1)

	list := []*mayType{}

	resp, err = ta.r().
		SetHeader("X-Text1-Read", "5678").
		SetResult(&list).
		Get("maytype")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 1)
	require.Equal(t, "5678", list[0].Text1)
}

func TestMayReadAPI(t *testing.T) {
	// TODO: Break up test
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ta.api.SetRequestHook(requestHook)
	api.Register[mayType](ta.api)

	created := &mayType{}

	resp, err := ta.r().
		SetBody(&mayType{Text1: "foo"}).
		SetResult(created).
		Post("maytype")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.NotEmpty(t, created.ID)

	get := &mayType{}

	resp, err = ta.r().
		SetHeader("X-NewText1", "abcd").
		SetResult(get).
		SetPathParam("id", created.ID).
		Get("maytype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "foo", get.Text1)

	list := []*mayType{}

	resp, err = ta.r().
		SetQueryParam("_sort", "+text1").
		SetResult(&list).
		Get("maytype")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 2)
	require.Equal(t, "abcd", list[0].Text1)
	require.Equal(t, "foo", list[1].Text1)
}
