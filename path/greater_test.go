package path

import "testing"
import "time"

import "cloud.google.com/go/civil"
import "github.com/stretchr/testify/require"

func TestGreaterInt(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		Int: 1234,
	}, "int", "1233")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		Int: 1234,
	}, "int", "1235")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterInt64(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		Int64: 3456,
	}, "int64", "3455")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		Int64: 3456,
	}, "int64", "3457")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterUInt(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		UInt: 4567,
	}, "uint", "4566")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		UInt: 4567,
	}, "uint", "4568")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterUInt64(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		UInt64: 5678,
	}, "uint64", "5677")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		UInt64: 5678,
	}, "uint64", "5679")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterFloat32(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		Float32: 3.1415,
	}, "float32", "3.1414")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		Float32: 3.1415,
	}, "float32", "3.1416")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterFloat64(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		Float64: 3.14159265,
	}, "float64", "3.14159264")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		Float64: 3.14159265,
	}, "float64", "3.14159266")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterString(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		String: "foo",
	}, "string", "bar")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		String: "foo",
	}, "string", "zig")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterBool(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		Bool: true,
	}, "bool", "false")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		Bool: false,
	}, "bool", "true")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterInts(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		Ints: []int{2, 4, 7},
	}, "ints", "5")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		Ints: []int{2, 4, 7},
	}, "ints", "8")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterInt64s(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		Int64s: []int64{2, 4, 7},
	}, "int64s", "5")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		Int64s: []int64{2, 4, 7},
	}, "int64s", "8")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterUInts(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		UInts: []uint{2, 4, 7},
	}, "uints", "5")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		UInts: []uint{2, 4, 7},
	}, "uints", "8")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterUInt64s(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		UInt64s: []uint64{2, 4, 7},
	}, "uint64s", "5")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		UInt64s: []uint64{2, 4, 7},
	}, "uint64s", "8")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterFloat32s(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		Float32s: []float32{3.1415, 2.7182},
	}, "float32s", "2.7181")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		Float32s: []float32{3.1415, 2.7182},
	}, "float32s", "3.1416")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterFloat64s(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		Float64s: []float64{3.1415, 2.7182},
	}, "float64s", "2.7181")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		Float64s: []float64{3.1415, 2.7182},
	}, "float64s", "3.1416")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterStrings(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		Strings: []string{"foo", "bar"},
	}, "strings", "baz")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		Strings: []string{"foo", "bar"},
	}, "strings", "zig")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterBools(t *testing.T) {
	t.Parallel()

	match, err := Greater(&testType1{
		Bools: []bool{true, false},
	}, "bools", "false")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		Bools: []bool{true, false},
	}, "bools", "true")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterTime(t *testing.T) {
	t.Parallel()

	tm, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	require.Nil(t, err)

	match, err := Greater(&testType1{
		Time: tm,
	}, "time", "2006-01-02T15:04:04Z")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		Time: tm,
	}, "time", "2006-01-02T15:04:06Z")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterTimes(t *testing.T) {
	t.Parallel()

	tm, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	require.Nil(t, err)

	tm2, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-10T15:04:05Z")
	require.Nil(t, err)

	match, err := Greater(&testType1{
		Times: []time.Time{tm, tm2},
	}, "times", "2006-01-05T15:04:05Z")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		Times: []time.Time{tm, tm2},
	}, "times", "2006-01-11T15:04:05Z")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterDate(t *testing.T) {
	t.Parallel()

	d, err := civil.ParseDate("2006-01-02")
	require.Nil(t, err)

	match, err := Greater(&testType1{
		Date: d,
	}, "date", "2006-01-01")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		Date: d,
	}, "date", "2006-01-03")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterDates(t *testing.T) {
	t.Parallel()

	d1, err := civil.ParseDate("2006-01-01")
	require.Nil(t, err)

	d2, err := civil.ParseDate("2006-01-03")
	require.Nil(t, err)

	match, err := Greater(&testType1{
		Dates: []civil.Date{d1, d2},
	}, "dates", "2006-01-02")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Greater(&testType1{
		Dates: []civil.Date{d1, d2},
	}, "dates", "2006-01-04")
	require.Nil(t, err)
	require.False(t, match)
}
