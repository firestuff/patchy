package api_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplateGoClient(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	gc, err := ta.pyc.GoClient(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, gc)
}

func TestTemplateTSClient(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	tc, err := ta.pyc.TSClient(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, tc)
}
