package api_test

import (
	"context"
	"testing"

	"github.com/firestuff/patchy/patchyc"
	"github.com/stretchr/testify/require"
)

func TestReplace(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo", Num: 1})
	require.NoError(t, err)

	replaced, err := patchyc.Replace(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, nil)
	require.NoError(t, err)
	require.NotNil(t, replaced)
	require.Equal(t, "bar", replaced.Text)
	require.EqualValues(t, 0, replaced.Num)
	require.EqualValues(t, created.Generation+1, replaced.Generation)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "bar", get.Text)
	require.EqualValues(t, 0, get.Num)
	require.EqualValues(t, created.Generation+1, get.Generation)
}

func TestReplaceNotExist(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	replaced, err := patchyc.Replace(ctx, ta.pyc, "doesnotexist", &testType{Text: "bar"}, nil)
	require.Error(t, err)
	require.Nil(t, replaced)
}

func TestReplaceIfMatchETagSuccess(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	replaced, err := patchyc.Replace(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, &patchyc.UpdateOpts{IfMatchETag: created.ETag})
	require.NoError(t, err)
	require.Equal(t, "bar", replaced.Text)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "bar", get.Text)
}

func TestReplaceIfMatchETagMismatch(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	replaced, err := patchyc.Replace(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, &patchyc.UpdateOpts{IfMatchETag: "etag:doesnotmatch"})
	require.Error(t, err)
	require.ErrorContains(t, err, "etag mismatch")
	require.Nil(t, replaced)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "foo", get.Text)
}

func TestReplaceIfMatchGenerationSuccess(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	replaced, err := patchyc.Replace(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, &patchyc.UpdateOpts{IfMatchGeneration: created.Generation})
	require.NoError(t, err)
	require.Equal(t, "bar", replaced.Text)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "bar", get.Text)
}

func TestReplaceIfMatchGenerationMismatch(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	replaced, err := patchyc.Replace(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, &patchyc.UpdateOpts{IfMatchGeneration: 50})
	require.Error(t, err)
	require.ErrorContains(t, err, "generation mismatch")
	require.Nil(t, replaced)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "foo", get.Text)
}

func TestReplaceIfMatchInvalid(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	ta.pyc.SetHeader("If-Match", `"foobar"`)

	replaced, err := patchyc.Replace(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid If-Match")
	require.Nil(t, replaced)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "foo", get.Text)
}
