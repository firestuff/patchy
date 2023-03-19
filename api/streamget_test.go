package api_test

import (
	"bufio"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStreamGet(t *testing.T) {
	// TODO: Break up test
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
	require.NoError(t, err)
	require.False(t, resp.IsError())

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetPathParam("id", created.ID).
		Get("testtype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	body := resp.RawBody()
	defer body.Close()

	scan := bufio.NewScanner(body)

	initial := &testType{}
	eventType, err := readEvent(scan, initial)
	require.NoError(t, err)
	require.Equal(t, "initial", eventType)
	require.Equal(t, "foo", initial.Text)

	// Heartbeat (after 5 seconds)
	eventType, err = readEvent(scan, nil)
	require.NoError(t, err)
	require.Equal(t, "heartbeat", eventType)

	updated := &testType{}

	// Round trip PATCH -> SSE
	resp, err = ta.r().
		SetBody(&testType{
			Text: "bar",
		}).
		SetResult(updated).
		SetPathParam("id", created.ID).
		Patch("testtype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	eventType, err = readEvent(scan, updated)
	require.NoError(t, err)
	require.Equal(t, "update", eventType)
	require.Equal(t, "bar", updated.Text)
}
