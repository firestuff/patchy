package api_test

import (
	"context"
	"testing"

	"github.com/firestuff/patchy/api"
	"github.com/firestuff/patchy/patchyc"
	"github.com/stretchr/testify/require"
)

func TestDirect(t *testing.T) {
	// TODO: Break up test
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	create, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)
	require.Equal(t, "foo", create.Text)

	get, err := api.Get[testType](ctx, ta.api, create.ID)
	require.NoError(t, err)
	require.Equal(t, create.ID, get.ID)
	require.Equal(t, "foo", get.Text)

	update, err := api.Update(ctx, ta.api, create.ID, &testType{Text: "bar"})
	require.NoError(t, err)
	require.Equal(t, create.ID, update.ID)
	require.Equal(t, "bar", update.Text)

	list, err := api.List[testType](ctx, ta.api, nil)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, create.ID, list[0].ID)
	require.Equal(t, "bar", list[0].Text)

	find, err := api.Find[testType](ctx, ta.api, create.ID[:4])
	require.NoError(t, err)
	require.Equal(t, create.ID, find.ID)
	require.Equal(t, "bar", find.Text)

	err = api.Delete[testType](ctx, ta.api, create.ID)
	require.NoError(t, err)

	list, err = api.List[testType](ctx, ta.api, nil)
	require.NoError(t, err)
	require.Len(t, list, 0)

	get, err = api.Get[testType](ctx, ta.api, "junk")
	require.NoError(t, err)
	require.Nil(t, get)
}

func TestDirectStreamGetNotFound(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	stream, err := api.StreamGet[testType](ctx, ta.api, "junk")
	require.Error(t, err)
	require.Nil(t, stream)
}

func TestDirectStreamGetInitial(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	create, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	stream, err := api.StreamGet[testType](ctx, ta.api, create.ID)
	require.NoError(t, err)
	require.NotNil(t, stream)

	defer stream.Close()

	obj := stream.Read()
	require.NotNil(t, obj)
	require.Equal(t, "foo", obj.Text)
}

func TestDirectStreamGetUpdate(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	create, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	stream, err := api.StreamGet[testType](ctx, ta.api, create.ID)
	require.NoError(t, err)
	require.NotNil(t, stream)

	defer stream.Close()

	obj := stream.Read()
	require.NotNil(t, obj)
	require.Equal(t, "foo", obj.Text)

	_, err = api.Update(ctx, ta.api, create.ID, &testType{Text: "bar"})
	require.NoError(t, err)

	obj = stream.Read()
	require.NotNil(t, obj)
	require.Equal(t, "bar", obj.Text)
}

func TestDirectStreamListInvalidType(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	stream, err := api.StreamListName[testType](ctx, ta.api, "invalid", nil)
	require.Error(t, err)
	require.Nil(t, stream)
}

func TestDirectStreamListInitial(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = api.Create(ctx, ta.api, &testType{Text: "bar"})
	require.NoError(t, err)

	stream, err := api.StreamList[testType](ctx, ta.api, nil)
	require.NoError(t, err)
	defer stream.Close()

	list := stream.Read()
	require.NotNil(t, list)
	require.Len(t, list, 2)
	require.ElementsMatch(t, []string{"foo", "bar"}, []string{list[0].Text, list[1].Text})
}

func TestDirectStreamListUpdate(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	stream, err := api.StreamList[testType](ctx, ta.api, nil)
	require.NoError(t, err)
	defer stream.Close()

	list := stream.Read()
	require.NotNil(t, list)
	require.Len(t, list, 0)

	_, err = api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	list = stream.Read()
	require.NotNil(t, list)
	require.Len(t, list, 1)
	require.Equal(t, "foo", list[0].Text)
}

func TestDirectStreamListDelete(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	stream, err := api.StreamList[testType](ctx, ta.api, nil)
	require.NoError(t, err)
	defer stream.Close()

	list := stream.Read()
	require.NotNil(t, list)
	require.Len(t, list, 1)
	require.Equal(t, "foo", list[0].Text)

	err = api.Delete[testType](ctx, ta.api, created.ID)
	require.NoError(t, err)

	list = stream.Read()
	require.NotNil(t, list)
	require.Len(t, list, 0)
}

func TestDirectStreamListOpts(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = api.Create(ctx, ta.api, &testType{Text: "bar"})
	require.NoError(t, err)

	stream, err := api.StreamList[testType](ctx, ta.api, &patchyc.ListOpts{Limit: 1})
	require.NoError(t, err)
	defer stream.Close()

	list := stream.Read()
	require.NotNil(t, list)
	require.Len(t, list, 1)
	require.Contains(t, []string{"foo", "bar"}, list[0].Text)
}
