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
	require.NoError(t, err)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.NotNil(t, get)
	require.Equal(t, "foo", get.Text)

	err = patchyc.Delete[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)

	get, err = patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Nil(t, get)
}

func TestDeleteInvalidID(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	err := patchyc.Delete[testType](ctx, ta.pyc, "doesnotexist")
	require.Error(t, err)
}

func TestDeleteTwice(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	err = patchyc.Delete[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)

	err = patchyc.Delete[testType](ctx, ta.pyc, created.ID)
	require.Error(t, err)
}

func TestDeleteStream(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	// TODO: Convert to using all patchyc once that supports streaming

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

	initial := &testType{}
	eventType, err := readEvent(scan, initial)
	require.NoError(t, err)
	require.Equal(t, "initial", eventType)
	require.Equal(t, "foo", initial.Text)

	err = patchyc.Delete[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)

	eventType, err = readEvent(scan, nil)
	require.NoError(t, err)
	require.Equal(t, "delete", eventType)
}
