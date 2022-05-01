package path_test

import (
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/firestuff/patchy/path"
	"github.com/stretchr/testify/require"
)

func TestInInt(t *testing.T) {
	t.Parallel()

	match, err := path.In(&testType1{
		Int: 1234,
	}, "int", "1233,1234,1235")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		Int: 1234,
	}, "int", "1233,1235")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInInt64(t *testing.T) {
	t.Parallel()

	match, err := path.In(&testType1{
		Int64: 3456,
	}, "int64", "3455,3456,3457")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		Int64: 3456,
	}, "int64", "3455,3457")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInUInt(t *testing.T) {
	t.Parallel()

	match, err := path.In(&testType1{
		UInt: 4567,
	}, "uint", "4566,4567,4568")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		UInt: 4567,
	}, "uint", "4566,4568")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInUInt64(t *testing.T) {
	t.Parallel()

	match, err := path.In(&testType1{
		UInt64: 5678,
	}, "uint64", "5677,5678,5679")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		UInt64: 5678,
	}, "uint64", "5677,5679")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInFloat32(t *testing.T) {
	t.Parallel()

	match, err := path.In(&testType1{
		Float32: 3.1415,
	}, "float32", "3.1414,3.1415,3.1416")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		Float32: 3.1415,
	}, "float32", "3.1414,3.1416")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInFloat64(t *testing.T) {
	t.Parallel()

	match, err := path.In(&testType1{
		Float64: 3.14159265,
	}, "float64", "3.14159264,3.14159265,3.14159266")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		Float64: 3.14159265,
	}, "float64", "3.14159264,3.14159266")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInString(t *testing.T) {
	t.Parallel()

	match, err := path.In(&testType1{
		String: "foo",
	}, "string2", "zig,foo,bar")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		String: "foo",
	}, "string2", "zig,bar")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInBool(t *testing.T) {
	t.Parallel()

	match, err := path.In(&testType1{
		Bool: true,
	}, "bool2", "true,false")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		Bool: true,
	}, "bool2", "false,false")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInInts(t *testing.T) {
	t.Parallel()

	match, err := path.In(&testType1{
		Ints: []int{2, 4, 7},
	}, "ints", "3,4,5")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		Ints: []int{2, 4, 7},
	}, "ints", "3,5")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInInt64s(t *testing.T) {
	t.Parallel()

	match, err := path.In(&testType1{
		Int64s: []int64{2, 4, 7},
	}, "int64s", "3,4,5")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		Int64s: []int64{2, 4, 7},
	}, "int64s", "3,5")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInUInts(t *testing.T) {
	t.Parallel()

	match, err := path.In(&testType1{
		UInts: []uint{2, 4, 7},
	}, "uints", "3,4,5")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		UInts: []uint{2, 4, 7},
	}, "uints", "3,5")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInUInt64s(t *testing.T) {
	t.Parallel()

	match, err := path.In(&testType1{
		UInt64s: []uint64{2, 4, 7},
	}, "uint64s", "3,4,5")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		UInt64s: []uint64{2, 4, 7},
	}, "uint64s", "3,5")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInFloat32s(t *testing.T) {
	t.Parallel()

	match, err := path.In(&testType1{
		Float32s: []float32{3.1415, 2.7182},
	}, "float32s", "2.7181,2.7182,2.7183")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		Float32s: []float32{3.1415, 2.7182},
	}, "float32s", "2.7181,2.7183")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInFloat64s(t *testing.T) {
	t.Parallel()

	match, err := path.In(&testType1{
		Float64s: []float64{3.1415, 2.7182},
	}, "float64s", "2.7181,2.7182,2.7183")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		Float64s: []float64{3.1415, 2.7182},
	}, "float64s", "2.7181,2.7183")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInStrings(t *testing.T) {
	t.Parallel()

	match, err := path.In(&testType1{
		Strings: []string{"foo", "bar"},
	}, "strings", "baz,foo,zig")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		Strings: []string{"foo", "bar"},
	}, "strings", "baz,zig")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInBools(t *testing.T) {
	t.Parallel()

	match, err := path.In(&testType1{
		Bools: []bool{true, false},
	}, "bools", "true,false")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		Bools: []bool{false, false},
	}, "bools", "true,true")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInTime(t *testing.T) {
	t.Parallel()

	tm, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	require.Nil(t, err)

	match, err := path.In(&testType1{
		Time: tm,
	}, "time", "2006-01-02T15:04:04Z,2006-01-02T15:04:05Z,2006-01-02T15:04:06Z")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		Time: tm,
	}, "time", "2006-01-02T15:04:04Z,2006-01-02T15:04:06Z")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInTimes(t *testing.T) {
	t.Parallel()

	tm, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	require.Nil(t, err)

	tm2, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-10T15:04:05Z")
	require.Nil(t, err)

	match, err := path.In(&testType1{
		Times: []time.Time{tm, tm2},
	}, "times", "2006-01-02T15:04:04Z,2006-01-02T15:04:05Z,2006-01-02T15:04:06Z")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		Times: []time.Time{tm, tm2},
	}, "times", "2006-01-02T15:04:04Z,2006-01-02T15:04:06Z")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInDate(t *testing.T) {
	t.Parallel()

	d, err := civil.ParseDate("2006-01-02")
	require.Nil(t, err)

	match, err := path.In(&testType1{
		Date: d,
	}, "date", "2006-01-01,2006-01-02,2006-01-03")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		Date: d,
	}, "date", "2006-01-01,2006-01-03")
	require.Nil(t, err)
	require.False(t, match)
}

func TestInDates(t *testing.T) {
	t.Parallel()

	d1, err := civil.ParseDate("2006-01-02")
	require.Nil(t, err)

	d2, err := civil.ParseDate("2006-01-05")
	require.Nil(t, err)

	match, err := path.In(&testType1{
		Dates: []civil.Date{d1, d2},
	}, "dates", "2006-01-01,2006-01-02,2006-01-03")
	require.Nil(t, err)
	require.True(t, match)

	match, err = path.In(&testType1{
		Dates: []civil.Date{d1, d2},
	}, "dates", "2006-01-01,2006-01-03")
	require.Nil(t, err)
	require.False(t, match)
}
