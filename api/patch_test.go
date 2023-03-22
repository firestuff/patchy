package api_test

import (
	"context"
	"testing"

	"github.com/firestuff/patchy"
	"github.com/firestuff/patchy/patchyc"
	"github.com/stretchr/testify/require"
)

func TestUpdate(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo", Num: 1})
	require.NoError(t, err)

	updated, err := patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, nil)
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, "bar", updated.Text)
	require.EqualValues(t, 1, updated.Num)
	require.EqualValues(t, created.Generation+1, updated.Generation)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "bar", get.Text)
	require.EqualValues(t, 1, get.Num)
	require.EqualValues(t, created.Generation+1, updated.Generation)
}

func TestUpdateNotExist(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	updated, err := patchyc.Update(ctx, ta.pyc, "doesnotexist", &testType{Text: "bar"}, nil)
	require.Error(t, err)
	require.Nil(t, updated)
}

func TestUpdateIfMatchETagSuccess(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	updated, err := patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, &patchyc.UpdateOpts{IfMatchETag: created.ETag})
	require.NoError(t, err)
	require.Equal(t, "bar", updated.Text)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "bar", get.Text)
}

func TestUpdateIfMatchETagMismatch(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	updated, err := patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, &patchy.UpdateOpts{IfMatchETag: "etag:doesnotmatch"})
	require.Error(t, err)
	require.ErrorContains(t, err, "etag mismatch")
	require.Nil(t, updated)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "foo", get.Text)
}

func TestUpdateIfMatchGenerationSuccess(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	updated, err := patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, &patchy.UpdateOpts{IfMatchGeneration: created.Generation})
	require.NoError(t, err)
	require.Equal(t, "bar", updated.Text)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "bar", get.Text)
}

func TestUpdateIfMatchGenerationMismatch(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	updated, err := patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, &patchy.UpdateOpts{IfMatchGeneration: 50})
	require.Error(t, err)
	require.ErrorContains(t, err, "generation mismatch")
	require.Nil(t, updated)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "foo", get.Text)
}

func TestUpdateIfMatchInvalid(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	ta.pyc.SetHeader("If-Match", `"foobar"`)

	updated, err := patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"}, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid If-Match")
	require.Nil(t, updated)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "foo", get.Text)
}
