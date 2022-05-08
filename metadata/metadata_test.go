package metadata_test

import (
	"testing"

	"github.com/firestuff/patchy/metadata"
	"github.com/stretchr/testify/require"
)

type TestType struct {
	metadata.Metadata
	Text string
}

func TestMetadata(t *testing.T) {
	t.Parallel()

	tt := &TestType{}

	// Verify promoted field
	tt.ID = "abc123"

	m := metadata.GetMetadata(tt)
	require.NotNil(t, m)
	require.Equal(t, "abc123", m.ID)

	metadata.ClearMetadata(tt)
	require.Empty(t, tt.ID)
}
