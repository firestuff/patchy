package api_test

import (
	"context"
	"testing"

	"github.com/firestuff/patchy/api"
	"github.com/stretchr/testify/require"
)

func TestDirect(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	create, err := api.Create(ctx, ta.api, &testType{Text: "foo"})
	require.Nil(t, err)
	require.Equal(t, "foo", create.Text)

	get, err := api.Get[testType](ctx, ta.api, create.ID)
	require.Nil(t, err)
	require.Equal(t, "foo", get.Text)
}
