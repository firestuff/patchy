package api_test

import (
	"context"
	"testing"

	"github.com/firestuff/patchy/api"
	"github.com/stretchr/testify/require"
)

func TestDirect(t *testing.T) {
	// TODO: Break up test
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	create, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.Nil(t, err)
	require.Equal(t, "foo", create.Text)

	get, err := api.Get[testType](ctx, ta.api, create.ID)
	require.Nil(t, err)
	require.Equal(t, create.ID, get.ID)
	require.Equal(t, "foo", get.Text)

	update, err := api.Update(ctx, ta.api, create.ID, &testType{Text: "bar"})
	require.Nil(t, err)
	require.Equal(t, create.ID, update.ID)
	require.Equal(t, "bar", update.Text)

	list, err := api.List[testType](ctx, ta.api, nil)
	require.Nil(t, err)
	require.Len(t, list, 1)
	require.Equal(t, create.ID, list[0].ID)
	require.Equal(t, "bar", list[0].Text)

	find, err := api.Find[testType](ctx, ta.api, create.ID[:4])
	require.Nil(t, err)
	require.Equal(t, create.ID, find.ID)
	require.Equal(t, "bar", find.Text)

	err = api.Delete[testType](ctx, ta.api, create.ID)
	require.Nil(t, err)

	list, err = api.List[testType](ctx, ta.api, nil)
	require.Nil(t, err)
	require.Len(t, list, 0)

	get, err = api.Get[testType](ctx, ta.api, "junk")
	require.Nil(t, err)
	require.Nil(t, get)
}
