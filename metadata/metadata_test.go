package metadata

import "testing"

import "github.com/stretchr/testify/require"

type TestType struct {
	Metadata
	Text string
}

func TestMetadata(t *testing.T) {
	t.Parallel()

	tt := &TestType{}

	// Verify promoted field
	tt.Id = "abc123"

	m := GetMetadata(tt)
	require.NotNil(t, m)
	require.Equal(t, "abc123", m.Id)
	require.Equal(t, "6ca13d52ca70c883e0f0bb101e425a89e8624de51db2d2392593af6a84118090", m.GetSafeId())

	ClearMetadata(tt)
	require.Empty(t, tt.Id)
}
