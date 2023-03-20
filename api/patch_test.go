package api_test

import (
	"context"
	"fmt"
	"testing"

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
	require.EqualValues(t, 0, created.Generation)

	updated, err := patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"})
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, "bar", updated.Text)
	require.EqualValues(t, 1, updated.Num)
	require.EqualValues(t, 1, updated.Generation)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "bar", get.Text)
	require.EqualValues(t, 1, get.Num)
	require.EqualValues(t, 1, updated.Generation)
}

func TestUpdateIfMatchETagSuccess(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create(ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	// TODO: Support If-Match directly in the client and direct APIs
	ta.pyc.SetHeader("If-Match", fmt.Sprintf(`"%s"`, created.ETag))

	updated, err := patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"})
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

	ta.pyc.SetHeader("If-Match", `"etag:doesnotmatch"`)

	updated, err := patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"})
	require.Error(t, err)
	require.ErrorContains(t, err, "[412] Precondition Failed")
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

	ta.pyc.SetHeader("If-Match", fmt.Sprintf(`"generation:%d"`, created.Generation))

	updated, err := patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"})
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

	ta.pyc.SetHeader("If-Match", `"generation:50"`)

	updated, err := patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"})
	require.Error(t, err)
	require.ErrorContains(t, err, "[412] Precondition Failed")
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

	updated, err := patchyc.Update(ctx, ta.pyc, created.ID, &testType{Text: "bar"})
	require.Error(t, err)
	require.ErrorContains(t, err, "[400] Bad Request")
	require.Nil(t, updated)

	get, err := patchyc.Get[testType](ctx, ta.pyc, created.ID)
	require.NoError(t, err)
	require.Equal(t, "foo", get.Text)
}
