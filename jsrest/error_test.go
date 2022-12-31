package jsrest_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/firestuff/patchy/jsrest"
	"github.com/stretchr/testify/require"
)

func TestFromError(t *testing.T) {
	t.Parallel()

	e1 := errors.New("error 1") //nolint:goerr113
	e2 := fmt.Errorf("error 2: %w", e1)

	e := jsrest.FromError(e2, jsrest.StatusBadGateway)

	require.Equal(t, e.Code, jsrest.StatusBadGateway)
	require.Equal(t, e.Messages, []string{"error 2: error 1", "error 1"})
}
