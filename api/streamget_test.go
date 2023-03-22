package api_test

import (
	"bufio"
	"context"
	"testing"

	"github.com/firestuff/patchy/api"
	"github.com/firestuff/patchy/patchyc"
	"github.com/stretchr/testify/require"
)

func TestStreamGetHeartbeat(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	resp, err := ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetPathParam("id", created.ID).
		Get("testtype/{id}")
	require.NoError(t, err)
	require.False(t, resp.IsError())

	body := resp.RawBody()
	defer body.Close()

	scan := bufio.NewScanner(body)

	eventType, _, err := readEvent[testType](scan)
	require.NoError(t, err)
	require.Equal(t, "initial", eventType)

	eventType, _, err = readEvent[api.EmptyEventType](scan)
	require.NoError(t, err)
	require.Equal(t, "heartbeat", eventType)
}

func TestStreamGet(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream, err := patchyc.StreamGet[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)

	defer stream.Close()

	get := stream.Read()
	require.NotNil(t, get)
	require.Equal(t, "foo", get.Text)
}

func TestStreamGetUpdate(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream, err := patchyc.StreamGet[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)

	defer stream.Close()

	get := stream.Read()
	require.NotNil(t, get)
	require.Equal(t, "foo", get.Text)

	_, err = patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, nil)
	require.NoError(t, err)

	updated := stream.Read()
	require.NotNil(t, updated)
	require.Equal(t, "bar", updated.Text)
}
