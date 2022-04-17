package path

import "testing"

import "github.com/stretchr/testify/require"

func TestPath(t *testing.T) {
	match, err := Match(&testType1{
		Int: 1234,
	}, "int", "1234")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Int: 1234,
	}, "int", "1235")
	require.Nil(t, err)
	require.False(t, match)

	match, err = Match(&testType2{
		Tt1: testType1{
			Int: 2345,
		},
	}, "tt1.int", "2345")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Int64: 3456,
	}, "int64", "3456")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		UInt: 4567,
	}, "uint", "4567")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		UInt64: 5678,
	}, "uint64", "5678")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		UInt64: 5678,
	}, "uint64", "5678")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		String: "foo",
	}, "string", "foo")
	require.Nil(t, err)
	require.True(t, match)
}

type testType1 struct {
	Int    int
	Int64  int64
	UInt   uint
	UInt64 uint64
	String string
}

type testType2 struct {
	Tt1 testType1
}
