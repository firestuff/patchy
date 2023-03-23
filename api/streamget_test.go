package api_test

import (
	"context"
	"testing"
	"time"

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

	stream, err := patchyc.StreamGet[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)

	defer stream.Close()

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Equal(t, "foo", ev.Obj.Text)

	time.Sleep(6 * time.Second)

	select {
	case _, ok := <-stream.Chan():
		if ok {
			require.Fail(t, "unexpected list")
		} else {
			require.Fail(t, "unexpected closure")
		}

	default:
	}

	require.Less(t, time.Since(stream.LastEventReceived()), 6*time.Second)
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

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Equal(t, "foo", ev.Obj.Text)
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

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Equal(t, "foo", ev.Obj.Text)

	_, err = patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, nil)
	require.NoError(t, err)

	ev = stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Equal(t, "bar", ev.Obj.Text)
}
