package api_test

import (
	"context"
	"testing"

	"github.com/firestuff/patchy/api"
	"github.com/firestuff/patchy/patchyc"
	"github.com/stretchr/testify/require"
)

func TestDirectGetNotFound(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	get, err := api.Get[testType](ctx, ta.api, "doesnotexist", nil)
	require.NoError(t, err)
	require.Nil(t, get)
}

func TestDirectGetInvalidType(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	create, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = api.GetName[testType](ctx, ta.api, "doesnotexist", create.ID, nil)
	require.Error(t, err)
}

func TestDirectCreateGet(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	create, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)
	require.Equal(t, "foo", create.Text)

	get, err := api.Get[testType](ctx, ta.api, create.ID, nil)
	require.NoError(t, err)
	require.Equal(t, create.ID, get.ID)
	require.Equal(t, "foo", get.Text)
}

func TestDirectCreateInvalidType(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := api.CreateName(ctx, ta.api, "doesnotexist", &testType{Text: "foo"})
	require.Error(t, err)
}

func TestDirectUpdate(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	create, err := api.Create(ctx, ta.api, &testType{Text: "foo", Num: 1})
	require.NoError(t, err)

	get, err := api.Get[testType](ctx, ta.api, create.ID, nil)
	require.NoError(t, err)
	require.Equal(t, "foo", get.Text)
	require.EqualValues(t, 1, get.Num)

	update, err := api.Update(ctx, ta.api, create.ID, &testType{Text: "bar"}, nil)
	require.NoError(t, err)
	require.Equal(t, create.ID, update.ID)
	require.Equal(t, "bar", update.Text)
	require.EqualValues(t, 1, update.Num)

	get, err = api.Get[testType](ctx, ta.api, create.ID, nil)
	require.NoError(t, err)
	require.Equal(t, "bar", get.Text)
	require.EqualValues(t, 1, get.Num)
}

func TestDirectUpdateInvalidType(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	create, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = api.UpdateName(ctx, ta.api, "doesnotexist", create.ID, &testType{Text: "bar"}, nil)
	require.Error(t, err)
}

func TestDirectReplace(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	create, err := api.Create(ctx, ta.api, &testType{Text: "foo", Num: 1})
	require.NoError(t, err)

	get, err := api.Get[testType](ctx, ta.api, create.ID, nil)
	require.NoError(t, err)
	require.Equal(t, "foo", get.Text)
	require.EqualValues(t, 1, get.Num)

	replace, err := api.Replace(ctx, ta.api, create.ID, &testType{Text: "bar"}, nil)
	require.NoError(t, err)
	require.Equal(t, create.ID, replace.ID)
	require.Equal(t, "bar", replace.Text)
	require.EqualValues(t, 0, replace.Num)

	get, err = api.Get[testType](ctx, ta.api, create.ID, nil)
	require.NoError(t, err)
	require.Equal(t, "bar", get.Text)
	require.EqualValues(t, 0, get.Num)
}

func TestDirectReplaceInvalidType(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	create, err := api.Create(ctx, ta.api, &testType{Text: "foo", Num: 1})
	require.NoError(t, err)

	get, err := api.Get[testType](ctx, ta.api, create.ID, nil)
	require.NoError(t, err)
	require.Equal(t, "foo", get.Text)
	require.EqualValues(t, 1, get.Num)

	_, err = api.ReplaceName(ctx, ta.api, "doesnotexist", create.ID, &testType{Text: "bar"}, nil)
	require.Error(t, err)
}

func TestDirectDelete(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	create, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = api.Get[testType](ctx, ta.api, create.ID, nil)
	require.NoError(t, err)

	err = api.Delete[testType](ctx, ta.api, create.ID, nil)
	require.NoError(t, err)

	get, err := api.Get[testType](ctx, ta.api, create.ID, nil)
	require.NoError(t, err)
	require.Nil(t, get)
}

func TestDirectDeleteInvalidType(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	create, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	err = api.DeleteName[testType](ctx, ta.api, "doesnotexist", create.ID, nil)
	require.Error(t, err)
}

func TestDirectDeleteNotFound(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	err := api.Delete[testType](ctx, ta.api, "doesnotexist", nil)
	require.Error(t, err)
}

func TestDirectList(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = api.Create(ctx, ta.api, &testType{Text: "bar"})
	require.NoError(t, err)

	list, err := api.List[testType](ctx, ta.api, nil)
	require.NoError(t, err)
	require.Len(t, list, 2)
	require.ElementsMatch(t, []string{"foo", "bar"}, []string{list[0].Text, list[1].Text})
}

func TestDirectListInvalidType(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = api.ListName[testType](ctx, ta.api, "doesnotexist", nil)
	require.Error(t, err)
}

func TestDirectFind(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	create, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	find, err := api.Find[testType](ctx, ta.api, create.ID[:4])
	require.NoError(t, err)
	require.Equal(t, create.ID, find.ID)
	require.Equal(t, "foo", find.Text)
}

func TestDirectFindNotExist(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	find, err := api.Find[testType](ctx, ta.api, "doesnotexist")
	require.Error(t, err)
	require.Nil(t, find)
}

func TestDirectFindMultiple(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = api.Create(ctx, ta.api, &testType{Text: "bar"})
	require.NoError(t, err)

	find, err := api.Find[testType](ctx, ta.api, "")
	require.Error(t, err)
	require.Nil(t, find)
}

func TestDirectStreamGetNotFound(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	stream, err := api.StreamGet[testType](ctx, ta.api, "junk")
	require.NoError(t, err)
	require.NotNil(t, stream)

	defer stream.Close()
}

func TestDirectStreamGetInvalidType(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	create, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = api.StreamGetName[testType](ctx, ta.api, "doesnotexist", create.ID)
	require.Error(t, err)
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

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Equal(t, "foo", ev.Obj.Text)
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

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Equal(t, "foo", ev.Obj.Text)

	_, err = api.Update(ctx, ta.api, create.ID, &testType{Text: "bar"}, nil)
	require.NoError(t, err)

	ev = stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Equal(t, "bar", ev.Obj.Text)
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

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 2)
	require.ElementsMatch(t, []string{"foo", "bar"}, []string{ev.List[0].Text, ev.List[1].Text})
}

func TestDirectStreamListUpdate(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	stream, err := api.StreamList[testType](ctx, ta.api, nil)
	require.NoError(t, err)

	defer stream.Close()

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 0)

	_, err = api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.NoError(t, err)

	ev = stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "foo", ev.List[0].Text)
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

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Equal(t, "foo", ev.List[0].Text)

	err = api.Delete[testType](ctx, ta.api, created.ID, nil)
	require.NoError(t, err)

	ev = stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 0)
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

	ev := stream.Read()
	require.NotNil(t, ev, stream.Error())
	require.Len(t, ev.List, 1)
	require.Contains(t, []string{"foo", "bar"}, ev.List[0].Text)
}
