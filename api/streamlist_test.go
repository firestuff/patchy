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

	eventType, err := readEvent(scan, nil)
	require.NoError(t, err)
	require.Equal(t, "list", eventType)

	eventType, err = readEvent(scan, nil)
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

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc)
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

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc)
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

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc)
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

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc)
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

/*

// TODO: Make StreamList() take ListOpts

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetQueryParam("_limit", "1").
		Get("testtype")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	body2 := resp.RawBody()
	defer body2.Close()

	scan2 := bufio.NewScanner(body2)

	eventType, err := readEvent(scan2, &list)
	require.NoError(t, err)
	require.Equal(t, "list", eventType)

	require.Len(t, list, 1)
	require.True(t, list[0].Text == "foo" || list[0].Text == "bar")
}
*/

func TestStreamListDiff(t *testing.T) {
	// TODO: Break up test
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	created1 := &testType{}

	resp, err := ta.r().
		SetBody(&testType{
			Text: "foo",
		}).
		SetResult(created1).
		Post("testtype")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	created2 := &testType{}

	resp, err = ta.r().
		SetBody(&testType{
			Text: "bar",
		}).
		SetResult(created2).
		Post("testtype")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetQueryParam("_stream", "diff").
		Get("testtype")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	body := resp.RawBody()
	defer body.Close()

	scan := bufio.NewScanner(body)

	resp2, err := ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetQueryParam("_stream", "diff").
		SetQueryParam("_sort", "text").
		SetQueryParam("_limit", "1").
		Get("testtype")
	require.NoError(t, err)
	require.False(t, resp2.IsError())

	body2 := resp2.RawBody()
	defer body2.Close()

	scan2 := bufio.NewScanner(body2)

	obj1 := testType{}

	eventType, err := readEvent(scan, &obj1)
	require.NoError(t, err)
	require.Equal(t, "add", eventType)

	obj2 := testType{}

	eventType, err = readEvent(scan, &obj2)
	require.NoError(t, err)
	require.Equal(t, "add", eventType)
	require.ElementsMatch(t, []string{"foo", "bar"}, []string{obj1.Text, obj2.Text})

	eventType, err = readEvent(scan2, &obj1)
	require.NoError(t, err)
	require.Equal(t, "add", eventType)
	require.Equal(t, "bar", obj1.Text)

	resp, err = ta.r().
		SetBody(&testType{
			Text: "zig",
		}).
		SetResult(created2).
		SetPathParam("id", created2.ID).
		Patch("testtype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	eventType, err = readEvent(scan, &obj1)
	require.NoError(t, err)
	require.Equal(t, "update", eventType)
	require.Equal(t, created2.ID, obj1.ID)
	require.Equal(t, "zig", obj1.Text)

	eventType, err = readEvent(scan2, &obj1)
	require.NoError(t, err)
	require.Equal(t, "add", eventType)
	require.Equal(t, created1.ID, obj1.ID)
	require.Equal(t, "foo", obj1.Text)

	eventType, err = readEvent(scan2, &obj1)
	require.NoError(t, err)
	require.Equal(t, "remove", eventType)
	require.Equal(t, created2.ID, obj1.ID)
	require.Equal(t, "bar", obj1.Text)

	resp, err = ta.r().
		SetPathParam("id", created1.ID).
		Delete("testtype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	eventType, err = readEvent(scan, &obj1)
	require.NoError(t, err)
	require.Equal(t, "remove", eventType)
	require.Equal(t, created1.ID, obj1.ID)
	require.Equal(t, "foo", obj1.Text)

	eventType, err = readEvent(scan2, &obj1)
	require.NoError(t, err)
	require.Equal(t, "add", eventType)
	require.Equal(t, created2.ID, obj1.ID)
	require.Equal(t, "zig", obj1.Text)

	eventType, err = readEvent(scan2, &obj1)
	require.NoError(t, err)
	require.Equal(t, "remove", eventType)
	require.Equal(t, created1.ID, obj1.ID)
	require.Equal(t, "foo", obj1.Text)
}
