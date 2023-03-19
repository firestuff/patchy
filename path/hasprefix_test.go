package path_test

import (
	"testing"

	"github.com/firestuff/patchy/path"
	"github.com/stretchr/testify/require"
)

func TestHasPrefixInt(t *testing.T) {
	t.Parallel()

	match, err := path.HasPrefix(&testType1{
		Int: -1234,
	}, "int", "-12")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.HasPrefix(&testType1{
		Int: -1234,
	}, "int", "23")
	require.NoError(t, err)
	require.False(t, match)
}

func TestHasPrefixInt64(t *testing.T) {
	t.Parallel()

	match, err := path.HasPrefix(&testType1{
		Int64: 3456,
	}, "int64", "34")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.HasPrefix(&testType1{
		Int64: 3456,
	}, "int64", "45")
	require.NoError(t, err)
	require.False(t, match)
}

func TestHasPrefixUInt(t *testing.T) {
	t.Parallel()

	match, err := path.HasPrefix(&testType1{
		UInt: 4567,
	}, "uint", "45")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.HasPrefix(&testType1{
		UInt: 4567,
	}, "uint", "457")
	require.NoError(t, err)
	require.False(t, match)
}

func TestHasPrefixUInt64(t *testing.T) {
	t.Parallel()

	match, err := path.HasPrefix(&testType1{
		UInt64: 5678,
	}, "uint64", "567")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.HasPrefix(&testType1{
		UInt64: 5678,
	}, "uint64", "56789")
	require.NoError(t, err)
	require.False(t, match)
}

func TestHasPrefixFloat32(t *testing.T) {
	t.Parallel()

	match, err := path.HasPrefix(&testType1{
		Float32: -3.1415,
	}, "float32", "-3.14")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.HasPrefix(&testType1{
		Float32: -3.1415,
	}, "float32", "3.14")
	require.NoError(t, err)
	require.False(t, match)
}

func TestHasPrefixFloat64(t *testing.T) {
	t.Parallel()

	match, err := path.HasPrefix(&testType1{
		Float64: 3.14159265,
	}, "float64", "3.1415926")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.HasPrefix(&testType1{
		Float64: 3.14159265,
	}, "float64", "3.141592651")
	require.NoError(t, err)
	require.False(t, match)
}

func TestHasPrefixString(t *testing.T) {
	t.Parallel()

	match, err := path.HasPrefix(&testType1{
		String: "foobar",
	}, "string2", "foo")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.HasPrefix(&testType1{
		String: "foobar",
	}, "string2", "bar")
	require.NoError(t, err)
	require.False(t, match)
}

func TestHasPrefixBool(t *testing.T) {
	t.Parallel()

	match, err := path.HasPrefix(&testType1{
		Bool: true,
	}, "bool2", "t")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.HasPrefix(&testType1{
		Bool: true,
	}, "bool2", "f")
	require.NoError(t, err)
	require.False(t, match)
}

func TestHasPrefixStrings(t *testing.T) {
	t.Parallel()

	match, err := path.HasPrefix(&testType1{
		Strings: []string{"foo", "bar"},
	}, "strings", "f")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.HasPrefix(&testType1{
		Strings: []string{"foo", "bar"},
	}, "strings", "z")
	require.NoError(t, err)
	require.False(t, match)
}
