package path_test

import (
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/firestuff/patchy/path"
	"github.com/stretchr/testify/require"
)

func TestLessInt(t *testing.T) {
	t.Parallel()

	match, err := path.Less(&testType1{
		Int: 1234,
	}, "int", "1235")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		Int: 1234,
	}, "int", "1233")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessInt64(t *testing.T) {
	t.Parallel()

	match, err := path.Less(&testType1{
		Int64: 3456,
	}, "int64", "3457")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		Int64: 3456,
	}, "int64", "3455")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessUInt(t *testing.T) {
	t.Parallel()

	match, err := path.Less(&testType1{
		UInt: 4567,
	}, "uint", "4568")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		UInt: 4567,
	}, "uint", "4566")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessUInt64(t *testing.T) {
	t.Parallel()

	match, err := path.Less(&testType1{
		UInt64: 5678,
	}, "uint64", "5679")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		UInt64: 5678,
	}, "uint64", "5677")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessFloat32(t *testing.T) {
	t.Parallel()

	match, err := path.Less(&testType1{
		Float32: 3.1415,
	}, "float32", "3.1416")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		Float32: 3.1415,
	}, "float32", "3.1414")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessFloat64(t *testing.T) {
	t.Parallel()

	match, err := path.Less(&testType1{
		Float64: 3.14159265,
	}, "float64", "3.14159266")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		Float64: 3.14159265,
	}, "float64", "3.14159264")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessString(t *testing.T) {
	t.Parallel()

	match, err := path.Less(&testType1{
		String: "foo",
	}, "string2", "zig")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		String: "foo",
	}, "string2", "bar")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessBool(t *testing.T) {
	t.Parallel()

	match, err := path.Less(&testType1{
		Bool: false,
	}, "bool2", "true")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		Bool: true,
	}, "bool2", "false")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessInts(t *testing.T) {
	t.Parallel()

	match, err := path.Less(&testType1{
		Ints: []int{2, 4, 7},
	}, "ints", "3")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		Ints: []int{2, 4, 7},
	}, "ints", "1")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessInt64s(t *testing.T) {
	t.Parallel()

	match, err := path.Less(&testType1{
		Int64s: []int64{2, 4, 7},
	}, "int64s", "3")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		Int64s: []int64{2, 4, 7},
	}, "int64s", "1")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessUInts(t *testing.T) {
	t.Parallel()

	match, err := path.Less(&testType1{
		UInts: []uint{2, 4, 7},
	}, "uints", "3")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		UInts: []uint{2, 4, 7},
	}, "uints", "1")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessUInt64s(t *testing.T) {
	t.Parallel()

	match, err := path.Less(&testType1{
		UInt64s: []uint64{2, 4, 7},
	}, "uint64s", "3")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		UInt64s: []uint64{2, 4, 7},
	}, "uint64s", "1")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessFloat32s(t *testing.T) {
	t.Parallel()

	match, err := path.Less(&testType1{
		Float32s: []float32{3.1415, 2.7182},
	}, "float32s", "3.1414")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		Float32s: []float32{3.1415, 2.7182},
	}, "float32s", "2.7181")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessFloat64s(t *testing.T) {
	t.Parallel()

	match, err := path.Less(&testType1{
		Float64s: []float64{3.1415, 2.7182},
	}, "float64s", "3.1414")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		Float64s: []float64{3.1415, 2.7182},
	}, "float64s", "2.7181")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessStrings(t *testing.T) {
	t.Parallel()

	match, err := path.Less(&testType1{
		Strings: []string{"foo", "bar"},
	}, "strings", "baz")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		Strings: []string{"foo", "bar"},
	}, "strings", "adv")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessBools(t *testing.T) {
	t.Parallel()

	match, err := path.Less(&testType1{
		Bools: []bool{true, false},
	}, "bools", "true")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		Bools: []bool{true, false},
	}, "bools", "false")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessTime(t *testing.T) {
	t.Parallel()

	tm, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	require.NoError(t, err)

	match, err := path.Less(&testType1{
		Time: tm,
	}, "time", "2006-01-02T15:04:06Z")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		Time: tm,
	}, "time", "2006-01-02T15:04:04Z")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessTimes(t *testing.T) {
	t.Parallel()

	tm, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	require.NoError(t, err)

	tm2, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-10T15:04:05Z")
	require.NoError(t, err)

	match, err := path.Less(&testType1{
		Times: []time.Time{tm, tm2},
	}, "times", "2006-01-05T15:04:05Z")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		Times: []time.Time{tm, tm2},
	}, "times", "2006-01-01T15:04:05Z")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessDate(t *testing.T) {
	t.Parallel()

	d, err := civil.ParseDate("2006-01-02")
	require.NoError(t, err)

	match, err := path.Less(&testType1{
		Date: d,
	}, "date", "2006-01-03")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		Date: d,
	}, "date", "2006-01-01")
	require.NoError(t, err)
	require.False(t, match)
}

func TestLessDates(t *testing.T) {
	t.Parallel()

	d1, err := civil.ParseDate("2006-01-02")
	require.NoError(t, err)

	d2, err := civil.ParseDate("2006-01-04")
	require.NoError(t, err)

	match, err := path.Less(&testType1{
		Dates: []civil.Date{d1, d2},
	}, "dates", "2006-01-03")
	require.NoError(t, err)
	require.True(t, match)

	match, err = path.Less(&testType1{
		Dates: []civil.Date{d1, d2},
	}, "dates", "2006-01-01")
	require.NoError(t, err)
	require.False(t, match)
}
