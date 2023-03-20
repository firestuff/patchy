package api_test

import (
	"bufio"
	"context"
	"testing"

	"github.com/firestuff/patchy/patchyc"
	"github.com/stretchr/testify/require"
)

func TestStreamListHeartbeat(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	resp, err := ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		Get("testtype")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	body := resp.RawBody()
	defer body.Close()

	scan := bufio.NewScanner(body)

	eventType, _, err := readEvent[[]*testType](scan)
	require.NoError(t, err)
	require.Equal(t, "list", eventType)

	eventType, _, err = readEvent[map[string]string](scan)
	require.NoError(t, err)
	require.Equal(t, "heartbeat", eventType)
}

func TestStreamListInitial(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = patchyc.Create(ctx, ta.pyc, &testType{Text: "bar"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, nil)
	require.NoError(t, err)

	defer stream.Close()

	list := stream.Read()
	require.NotNil(t, list)
	require.Len(t, list, 2)
	require.ElementsMatch(t, []string{"foo", "bar"}, []string{list[0].Text, list[1].Text})
}

func TestStreamListAdd(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, nil)
	require.NoError(t, err)

	defer stream.Close()

	list := stream.Read()
	require.NotNil(t, list)
	require.Len(t, list, 0)

	_, err = patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	list = stream.Read()
	require.NotNil(t, list)
	require.Len(t, list, 1)
	require.Equal(t, "foo", list[0].Text)
}

func TestStreamListUpdate(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, nil)
	require.NoError(t, err)

	defer stream.Close()

	list := stream.Read()
	require.NotNil(t, list)
	require.Len(t, list, 1)
	require.Equal(t, "foo", list[0].Text)

	_, err = patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"})
	require.NoError(t, err)

	list = stream.Read()
	require.NotNil(t, list)
	require.Len(t, list, 1)
	require.Equal(t, "bar", list[0].Text)
}

func TestStreamListDelete(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, nil)
	require.NoError(t, err)

	defer stream.Close()

	list := stream.Read()
	require.NotNil(t, list)
	require.Len(t, list, 1)
	require.Equal(t, "foo", list[0].Text)

	err = patchyc.Delete[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)

	list = stream.Read()
	require.NotNil(t, list)
	require.Len(t, list, 0)
}

func TestStreamListOpts(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = patchyc.Create(ctx, ta.pyc, &testType{Text: "bar"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Limit: 1})
	require.NoError(t, err)

	defer stream.Close()

	list := stream.Read()
	require.NotNil(t, list)
	require.Len(t, list, 1)
	require.Contains(t, []string{"foo", "bar"}, list[0].Text)
}

func TestStreamListDiffInitial(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	// No client support for this, use direct queries
	resp, err := ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetQueryParam("_stream", "diff").
		Get("testtype")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	body := resp.RawBody()
	defer body.Close()

	scan := bufio.NewScanner(body)

	eventType, obj, err := readEvent[testType](scan)
	require.NoError(t, err)
	require.Equal(t, "add", eventType)
	require.Equal(t, created.ID, obj.ID)
	require.Equal(t, "foo", obj.Text)
}

func TestStreamListDiffCreate(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	resp, err := ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetQueryParam("_stream", "diff").
		Get("testtype")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	body := resp.RawBody()
	defer body.Close()

	scan := bufio.NewScanner(body)

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	eventType, obj, err := readEvent[testType](scan)
	require.NoError(t, err)
	require.Equal(t, "add", eventType)
	require.Equal(t, created.ID, obj.ID)
	require.Equal(t, "foo", obj.Text)
}

func TestStreamListDiffUpdate(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo", Num: 1})
	require.NoError(t, err)

	resp, err := ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetQueryParam("_stream", "diff").
		Get("testtype")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	body := resp.RawBody()
	defer body.Close()

	scan := bufio.NewScanner(body)

	eventType, obj, err := readEvent[testType](scan)
	require.NoError(t, err)
	require.Equal(t, "add", eventType)
	require.Equal(t, created.ID, obj.ID)
	require.Equal(t, "foo", obj.Text)
	require.EqualValues(t, 1, obj.Num)

	_, err = patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"})
	require.NoError(t, err)

	eventType, obj, err = readEvent[testType](scan)
	require.NoError(t, err)
	require.Equal(t, "update", eventType)
	require.Equal(t, created.ID, obj.ID)
	require.Equal(t, "bar", obj.Text)
	require.EqualValues(t, 1, obj.Num)
}

func TestStreamListDiffReplace(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo", Num: 1})
	require.NoError(t, err)

	resp, err := ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetQueryParam("_stream", "diff").
		Get("testtype")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	body := resp.RawBody()
	defer body.Close()

	scan := bufio.NewScanner(body)

	eventType, obj, err := readEvent[testType](scan)
	require.NoError(t, err)
	require.Equal(t, "add", eventType)
	require.Equal(t, created.ID, obj.ID)
	require.Equal(t, "foo", obj.Text)
	require.EqualValues(t, 1, obj.Num)

	_, err = patchyc.Replace(ctx, ta.pyc, created.ID, &testType{Text: "bar"})
	require.NoError(t, err)

	eventType, obj, err = readEvent[testType](scan)
	require.NoError(t, err)
	require.Equal(t, "update", eventType)
	require.Equal(t, created.ID, obj.ID)
	require.Equal(t, "bar", obj.Text)
	require.EqualValues(t, 0, obj.Num)
}

func TestStreamListDiffDelete(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	resp, err := ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetQueryParam("_stream", "diff").
		Get("testtype")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	body := resp.RawBody()
	defer body.Close()

	scan := bufio.NewScanner(body)

	eventType, obj, err := readEvent[testType](scan)
	require.NoError(t, err)
	require.Equal(t, "add", eventType)
	require.Equal(t, created.ID, obj.ID)
	require.Equal(t, "foo", obj.Text)

	err = patchyc.Delete[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)

	eventType, obj, err = readEvent[testType](scan)
	require.NoError(t, err)
	require.Equal(t, "remove", eventType)
	require.Equal(t, created.ID, obj.ID)
	require.Equal(t, "foo", obj.Text)
}

func TestStreamListDiffSort(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created1, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	resp, err := ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetQueryParam("_stream", "diff").
		SetQueryParam("_sort", "text").
		SetQueryParam("_limit", "1").
		Get("testtype")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	body := resp.RawBody()
	defer body.Close()

	scan := bufio.NewScanner(body)

	eventType, obj, err := readEvent[testType](scan)
	require.NoError(t, err)
	require.Equal(t, "add", eventType)
	require.Equal(t, created1.ID, obj.ID)
	require.Equal(t, "foo", obj.Text)

	created2, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "bar"})
	require.NoError(t, err)

	eventType, obj, err = readEvent[testType](scan)
	require.NoError(t, err)
	require.Equal(t, "add", eventType)
	require.Equal(t, created2.ID, obj.ID)
	require.Equal(t, "bar", obj.Text)

	eventType, obj, err = readEvent[testType](scan)
	require.NoError(t, err)
	require.Equal(t, "remove", eventType)
	require.Equal(t, created1.ID, obj.ID)
	require.Equal(t, "foo", obj.Text)
}
