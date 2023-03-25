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

	stream, err := patchyc.StreamGet[testType](ctx, ta.pyc, created.ID, nil)
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

	stream, err := patchyc.StreamGet[testType](ctx, ta.pyc, created.ID, nil)
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

	stream, err := patchyc.StreamGet[testType](ctx, ta.pyc, created.ID, nil)
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

func TestStreamGetPrev(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream1, err := patchyc.StreamGet[testType](ctx, ta.pyc, created.ID, nil)
	require.NoError(t, err)

	defer stream1.Close()

	ev := stream1.Read()
	require.NotNil(t, ev, stream1.Error())
	require.Equal(t, "foo", ev.Obj.Text)
	require.EqualValues(t, 0, ev.Obj.Num)

	// Validate that previous version passing only compares the ETag
	ev.Obj.Num = 1

	stream2, err := patchyc.StreamGet[testType](ctx, ta.pyc, created.ID, &patchyc.GetOpts{Prev: ev.Obj})
	require.NoError(t, err)

	defer stream2.Close()

	ev2 := stream2.Read()
	require.NotNil(t, ev2, stream2.Error())
	require.Equal(t, "foo", ev2.Obj.Text)
	require.EqualValues(t, 1, ev2.Obj.Num)
}
