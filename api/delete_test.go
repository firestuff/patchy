package api_test

import (
	"bufio"
	"context"
	"testing"

	"github.com/firestuff/patchy/patchyc"
	"github.com/stretchr/testify/require"
)

func TestDeleteSuccess(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.Nil(t, err)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.Nil(t, err)
	require.NotNil(t, get)
	require.Equal(t, "foo", get.Text)

	err = patchyc.Delete[testType](ctx, ta.pyc, created.ID)
	require.Nil(t, err)

	get, err = patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.Nil(t, err)
	require.Nil(t, get)
}

func TestDeleteStream(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.Nil(t, err)

	// TODO: Convert to using all patchyc once that supports streaming

	resp, err := ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetPathParam("id", created.ID).
		Get("testtype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	body := resp.RawBody()
	defer body.Close()

	scan := bufio.NewScanner(body)

	initial := &testType{}
	eventType, err := readEvent(scan, initial)
	require.Nil(t, err)
	require.Equal(t, "initial", eventType)
	require.Equal(t, "foo", initial.Text)

	err = patchyc.Delete[testType](ctx, ta.pyc, created.ID)
	require.Nil(t, err)

	eventType, err = readEvent(scan, nil)
	require.Nil(t, err)
	require.Equal(t, "delete", eventType)
}
