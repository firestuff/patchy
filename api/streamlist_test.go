package api_test

import (
	"bufio"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStreamList(t *testing.T) {
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
	require.Nil(t, err)
	require.False(t, resp.IsError())

	created2 := &testType{}

	resp, err = ta.r().
		SetBody(&testType{
			Text: "bar",
		}).
		SetResult(created2).
		Post("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	body := resp.RawBody()
	defer body.Close()

	scan := bufio.NewScanner(body)

	list := []*testType{}

	eventType, err := readEvent(scan, &list)
	require.Nil(t, err)
	require.Equal(t, "list", eventType)

	require.Len(t, list, 2)
	require.ElementsMatch(t, []string{"foo", "bar"}, []string{list[0].Text, list[1].Text})

	// Heartbeat (after 5 seconds)
	eventType, err = readEvent(scan, nil)
	require.Nil(t, err)
	require.Equal(t, "heartbeat", eventType)

	created3 := &testType{}

	resp, err = ta.r().
		SetBody(&testType{
			Text: "zig",
		}).
		SetResult(created3).
		Post("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	eventType, err = readEvent(scan, &list)
	require.Nil(t, err)
	require.Equal(t, "list", eventType)

	require.Len(t, list, 3)
	require.ElementsMatch(t, []string{"foo", "bar", "zig"}, []string{list[0].Text, list[1].Text, list[2].Text})

	resp, err = ta.r().
		SetBody(&testType{
			Text: "zag",
		}).
		SetResult(created3).
		SetPathParam("id", created3.ID).
		Patch("testtype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	eventType, err = readEvent(scan, &list)
	require.Nil(t, err)
	require.Equal(t, "list", eventType)

	require.Len(t, list, 3)
	require.ElementsMatch(t, []string{"foo", "bar", "zag"}, []string{list[0].Text, list[1].Text, list[2].Text})

	resp, err = ta.r().
		SetPathParam("id", created3.ID).
		Delete("testtype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	eventType, err = readEvent(scan, &list)
	require.Nil(t, err)
	require.Equal(t, "list", eventType)

	require.Len(t, list, 2)
	require.ElementsMatch(t, []string{"foo", "bar"}, []string{list[0].Text, list[1].Text})

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetQueryParam("_limit", "1").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	body2 := resp.RawBody()
	defer body2.Close()

	scan2 := bufio.NewScanner(body2)

	eventType, err = readEvent(scan2, &list)
	require.Nil(t, err)
	require.Equal(t, "list", eventType)

	require.Len(t, list, 1)
	require.True(t, list[0].Text == "foo" || list[0].Text == "bar")
}

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
	require.Nil(t, err)
	require.False(t, resp.IsError())

	created2 := &testType{}

	resp, err = ta.r().
		SetBody(&testType{
			Text: "bar",
		}).
		SetResult(created2).
		Post("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetQueryParam("_stream", "diff").
		Get("testtype")
	require.Nil(t, err)
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
	require.Nil(t, err)
	require.False(t, resp2.IsError())

	body2 := resp2.RawBody()
	defer body2.Close()

	scan2 := bufio.NewScanner(body2)

	obj1 := testType{}

	eventType, err := readEvent(scan, &obj1)
	require.Nil(t, err)
	require.Equal(t, "add", eventType)

	obj2 := testType{}

	eventType, err = readEvent(scan, &obj2)
	require.Nil(t, err)
	require.Equal(t, "add", eventType)
	require.ElementsMatch(t, []string{"foo", "bar"}, []string{obj1.Text, obj2.Text})

	eventType, err = readEvent(scan2, &obj1)
	require.Nil(t, err)
	require.Equal(t, "add", eventType)
	require.Equal(t, "bar", obj1.Text)

	resp, err = ta.r().
		SetBody(&testType{
			Text: "zig",
		}).
		SetResult(created2).
		SetPathParam("id", created2.ID).
		Patch("testtype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	eventType, err = readEvent(scan, &obj1)
	require.Nil(t, err)
	require.Equal(t, "update", eventType)
	require.Equal(t, created2.ID, obj1.ID)
	require.Equal(t, "zig", obj1.Text)

	eventType, err = readEvent(scan2, &obj1)
	require.Nil(t, err)
	require.Equal(t, "add", eventType)
	require.Equal(t, created1.ID, obj1.ID)
	require.Equal(t, "foo", obj1.Text)

	eventType, err = readEvent(scan2, &obj1)
	require.Nil(t, err)
	require.Equal(t, "remove", eventType)
	require.Equal(t, created2.ID, obj1.ID)
	require.Equal(t, "bar", obj1.Text)

	resp, err = ta.r().
		SetPathParam("id", created1.ID).
		Delete("testtype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	eventType, err = readEvent(scan, &obj1)
	require.Nil(t, err)
	require.Equal(t, "remove", eventType)
	require.Equal(t, created1.ID, obj1.ID)
	require.Equal(t, "foo", obj1.Text)

	eventType, err = readEvent(scan2, &obj1)
	require.Nil(t, err)
	require.Equal(t, "add", eventType)
	require.Equal(t, created2.ID, obj1.ID)
	require.Equal(t, "zig", obj1.Text)

	eventType, err = readEvent(scan2, &obj1)
	require.Nil(t, err)
	require.Equal(t, "remove", eventType)
	require.Equal(t, created1.ID, obj1.ID)
	require.Equal(t, "foo", obj1.Text)
}
