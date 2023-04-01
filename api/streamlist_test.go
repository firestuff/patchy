package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/firestuff/patchy/api"
	"github.com/firestuff/patchy/patchyc"
	"github.com/stretchr/testify/require"
)

func TestStreamListHeartbeat(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, nil)
	require.NoError(t, err)

	defer stream.Close()

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 0)

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

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 2)
	require.ElementsMatch(t, []string{"foo", "bar"}, []string{ev.List[0].Text, ev.List[1].Text})
}

func TestStreamListAdd(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, nil)
	require.NoError(t, err)

	defer stream.Close()

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 0)

	_, err = patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	ev = stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "foo", ev.List[0].Text)
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

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "foo", ev.List[0].Text)

	_, err = patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, nil)
	require.NoError(t, err)

	ev = stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "bar", ev.List[0].Text)
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

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "foo", ev.List[0].Text)

	err = patchyc.Delete[testType](ctx, ta.pyc, created.ID, nil)
	require.NoError(t, err)

	ev = stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 0)
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

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Contains(t, []string{"foo", "bar"}, ev.List[0].Text)
}

func TestStreamListIgnoreIrrelevant(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created1, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = patchyc.Create(ctx, ta.pyc, &testType{Text: "bar"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Sorts: []string{"+text"}, Limit: 1})
	require.NoError(t, err)

	defer stream.Close()

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "bar", ev.List[0].Text)

	_, err = patchyc.Update(ctx, ta.pyc, created1.ID, &testType{Text: "zig"}, nil)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	select {
	case ev := <-stream.Chan():
		require.Fail(t, "unexpected update", ev)
	default:
	}

	_, err = patchyc.Update(ctx, ta.pyc, created1.ID, &testType{Text: "aaa"}, nil)
	require.NoError(t, err)

	ev = stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "aaa", ev.List[0].Text)
}

func TestStreamListPrev(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream1, err := patchyc.StreamList[testType](ctx, ta.pyc, nil)
	require.NoError(t, err)

	defer stream1.Close()

	ev := stream1.Read()
	require.NotNil(t, ev, stream1.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "foo", ev.List[0].Text)

	stream2, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Prev: ev.List})
	require.NoError(t, err)

	defer stream2.Close()

	ev = stream2.Read()
	require.NotNil(t, ev, stream2.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "foo", ev.List[0].Text)
}

func TestStreamListDiffInitial(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff"})
	require.NoError(t, err)

	defer stream.Close()

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "foo", ev.List[0].Text)
}

func TestStreamListDiffCreate(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff"})
	require.NoError(t, err)

	defer stream.Close()

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 0)

	_, err = patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	ev = stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "foo", ev.List[0].Text)
}

func TestStreamListDiffUpdate(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo", Num: 1})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff"})
	require.NoError(t, err)

	defer stream.Close()

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "foo", ev.List[0].Text)
	require.EqualValues(t, 1, ev.List[0].Num)

	_, err = patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, nil)
	require.NoError(t, err)

	ev = stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "bar", ev.List[0].Text)
	require.EqualValues(t, 1, ev.List[0].Num)
}

func TestStreamListDiffReplace(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo", Num: 1})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff"})
	require.NoError(t, err)

	defer stream.Close()

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "foo", ev.List[0].Text)
	require.EqualValues(t, 1, ev.List[0].Num)

	_, err = patchyc.Replace(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, nil)
	require.NoError(t, err)

	ev = stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "bar", ev.List[0].Text)
	require.EqualValues(t, 0, ev.List[0].Num)
}

func TestStreamListDiffDelete(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff"})
	require.NoError(t, err)

	defer stream.Close()

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "foo", ev.List[0].Text)

	err = patchyc.Delete[testType](ctx, ta.pyc, created.ID, nil)
	require.NoError(t, err)

	ev = stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 0)
}

func TestStreamListDiffSort(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{
		Stream: "diff",
		Sorts:  []string{"text"},
		Limit:  1,
	})
	require.NoError(t, err)

	defer stream.Close()

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "foo", ev.List[0].Text)

	_, err = patchyc.Create(ctx, ta.pyc, &testType{Text: "bar"})
	require.NoError(t, err)

	ev = stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "bar", ev.List[0].Text)
}

func TestStreamListDiffIgnoreIrrelevant(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created1, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = patchyc.Create(ctx, ta.pyc, &testType{Text: "bar"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff", Sorts: []string{"+text"}, Limit: 1})
	require.NoError(t, err)

	defer stream.Close()

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "bar", ev.List[0].Text)

	_, err = patchyc.Update(ctx, ta.pyc, created1.ID, &testType{Text: "zig"}, nil)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	select {
	case ev := <-stream.Chan():
		require.Fail(t, "unexpected update", ev)
	default:
	}

	_, err = patchyc.Update(ctx, ta.pyc, created1.ID, &testType{Text: "aaa"}, nil)
	require.NoError(t, err)

	ev = stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "aaa", ev.List[0].Text)
}

func TestStreamListDiffPrev(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream1, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff"})
	require.NoError(t, err)

	defer stream1.Close()

	ev := stream1.Read()
	require.NotNil(t, ev, stream1.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "foo", ev.List[0].Text)

	stream2, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff", Prev: ev.List})
	require.NoError(t, err)

	defer stream2.Close()

	ev = stream2.Read()
	require.NotNil(t, ev, stream2.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "foo", ev.List[0].Text)
}

func TestStreamListDefaultDiff(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ta.api.SetDefaultListOpts(&api.ListOpts{
		Stream: "diff",
	})

	ctx := context.Background()

	_, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = patchyc.Create(ctx, ta.pyc, &testType{Text: "bar"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, nil)
	require.NoError(t, err)

	defer stream.Close()

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 2)
	require.ElementsMatch(t, []string{"foo", "bar"}, []string{ev.List[0].Text, ev.List[1].Text})
}
