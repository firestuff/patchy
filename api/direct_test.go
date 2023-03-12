package api_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDirect(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	create, err := ta.api.Create(ctx, "testtype", &testType{Text: "foo"})
	require.Nil(t, err)
	require.Equal(t, "foo", create.(*testType).Text)

	get, err := ta.api.Get(ctx, "testtype", create.(*testType).ID)
	require.Nil(t, err)
	require.Equal(t, "foo", get.(*testType).Text)
}
