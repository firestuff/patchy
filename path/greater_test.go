package path

import "testing"

import "github.com/stretchr/testify/require"

func TestGreaterInt(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		Int: 1234,
	}, "int", "1233")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Int: 1234,
	}, "int", "1235")
	require.Nil(t, err)
	require.False(t, match)
}
