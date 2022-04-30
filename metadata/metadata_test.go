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
	require.Equal(t, "6ca13d52ca70c883e0f0bb101e425a89e8624de51db2d2392593af6a84118090", m.GetSafeID())

	metadata.ClearMetadata(tt)
	require.Empty(t, tt.ID)
}
