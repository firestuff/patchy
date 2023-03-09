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

func TestFromErrors(t *testing.T) {
	t.Parallel()

	e1 := errors.New("error 1") //nolint:goerr113
	e2 := errors.New("error 2") //nolint:goerr113
	e3 := fmt.Errorf("error 3: %w + %w", e1, e2)

	e := jsrest.FromError(e3, jsrest.StatusBadGateway)

	require.Equal(t, e.Code, jsrest.StatusBadGateway)
	require.Equal(t, e.Messages, []string{"error 3: error 1 + error 2", "error 1", "error 2"})
}

func TestParams(t *testing.T) {
	t.Parallel()

	e1 := errors.New("error 1") //nolint:goerr113

	e := jsrest.FromError(e1, jsrest.StatusUnauthorized)
	e.SetParam("foo", "bar")

	require.Contains(t, e.Error(), `"error 1"`)
	require.Contains(t, e.Error(), `"foo":`)
	require.Contains(t, e.Error(), `"bar"`)
}
