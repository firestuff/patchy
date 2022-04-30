package path

import (
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/stretchr/testify/require"
)

func TestGreaterEqualInt(t *testing.T) {
	t.Parallel()

	match, err := GreaterEqual(&testType1{
		Int: 1234,
	}, "int", "1233")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Int: 1234,
	}, "int", "1234")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Int: 1234,
	}, "int", "1235")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualInt64(t *testing.T) {
	t.Parallel()

	match, err := GreaterEqual(&testType1{
		Int64: 3456,
	}, "int64", "3455")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Int64: 3456,
	}, "int64", "3456")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Int64: 3456,
	}, "int64", "3457")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualUInt(t *testing.T) {
	t.Parallel()

	match, err := GreaterEqual(&testType1{
		UInt: 4567,
	}, "uint", "4566")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		UInt: 4567,
	}, "uint", "4567")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		UInt: 4567,
	}, "uint", "4568")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualUInt64(t *testing.T) {
	t.Parallel()

	match, err := GreaterEqual(&testType1{
		UInt64: 5678,
	}, "uint64", "5677")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		UInt64: 5678,
	}, "uint64", "5678")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		UInt64: 5678,
	}, "uint64", "5679")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualFloat32(t *testing.T) {
	t.Parallel()

	match, err := GreaterEqual(&testType1{
		Float32: 3.1415,
	}, "float32", "3.1414")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Float32: 3.1415,
	}, "float32", "3.1415")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Float32: 3.1415,
	}, "float32", "3.1416")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualFloat64(t *testing.T) {
	t.Parallel()

	match, err := GreaterEqual(&testType1{
		Float64: 3.14159265,
	}, "float64", "3.14159264")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Float64: 3.14159265,
	}, "float64", "3.14159265")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Float64: 3.14159265,
	}, "float64", "3.14159266")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualString(t *testing.T) {
	t.Parallel()

	match, err := GreaterEqual(&testType1{
		String: "foo",
	}, "string2", "bar")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		String: "foo",
	}, "string2", "foo")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		String: "foo",
	}, "string2", "zig")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualBool(t *testing.T) {
	t.Parallel()

	match, err := GreaterEqual(&testType1{
		Bool: true,
	}, "bool2", "false")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Bool: true,
	}, "bool2", "true")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Bool: false,
	}, "bool2", "true")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualInts(t *testing.T) {
	t.Parallel()

	match, err := GreaterEqual(&testType1{
		Ints: []int{2, 4, 7},
	}, "ints", "5")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Ints: []int{2, 4, 7},
	}, "ints", "7")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Ints: []int{2, 4, 7},
	}, "ints", "8")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualInt64s(t *testing.T) {
	t.Parallel()

	match, err := GreaterEqual(&testType1{
		Int64s: []int64{2, 4, 7},
	}, "int64s", "5")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Int64s: []int64{2, 4, 7},
	}, "int64s", "7")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Int64s: []int64{2, 4, 7},
	}, "int64s", "8")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualUInts(t *testing.T) {
	t.Parallel()

	match, err := GreaterEqual(&testType1{
		UInts: []uint{2, 4, 7},
	}, "uints", "5")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		UInts: []uint{2, 4, 7},
	}, "uints", "7")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		UInts: []uint{2, 4, 7},
	}, "uints", "8")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualUInt64s(t *testing.T) {
	t.Parallel()

	match, err := GreaterEqual(&testType1{
		UInt64s: []uint64{2, 4, 7},
	}, "uint64s", "5")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		UInt64s: []uint64{2, 4, 7},
	}, "uint64s", "7")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		UInt64s: []uint64{2, 4, 7},
	}, "uint64s", "8")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualFloat32s(t *testing.T) {
	t.Parallel()

	match, err := GreaterEqual(&testType1{
		Float32s: []float32{3.1415, 2.7182},
	}, "float32s", "2.7181")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Float32s: []float32{3.1415, 2.7182},
	}, "float32s", "3.1415")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Float32s: []float32{3.1415, 2.7182},
	}, "float32s", "3.1416")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualFloat64s(t *testing.T) {
	t.Parallel()

	match, err := GreaterEqual(&testType1{
		Float64s: []float64{3.1415, 2.7182},
	}, "float64s", "2.7181")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Float64s: []float64{3.1415, 2.7182},
	}, "float64s", "3.1415")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Float64s: []float64{3.1415, 2.7182},
	}, "float64s", "3.1416")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualStrings(t *testing.T) {
	t.Parallel()

	match, err := GreaterEqual(&testType1{
		Strings: []string{"foo", "bar"},
	}, "strings", "baz")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Strings: []string{"foo", "bar"},
	}, "strings", "foo")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Strings: []string{"foo", "bar"},
	}, "strings", "zig")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualBools(t *testing.T) {
	t.Parallel()

	match, err := GreaterEqual(&testType1{
		Bools: []bool{true, false},
	}, "bools", "false")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Bools: []bool{true, false},
	}, "bools", "true")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Bools: []bool{false, false},
	}, "bools", "true")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualTime(t *testing.T) {
	t.Parallel()

	tm, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	require.Nil(t, err)

	match, err := GreaterEqual(&testType1{
		Time: tm,
	}, "time", "2006-01-02T15:04:04Z")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Time: tm,
	}, "time", "2006-01-02T15:04:05Z")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Time: tm,
	}, "time", "2006-01-02T15:04:06Z")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualTimes(t *testing.T) {
	t.Parallel()

	tm, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	require.Nil(t, err)

	tm2, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-10T15:04:05Z")
	require.Nil(t, err)

	match, err := GreaterEqual(&testType1{
		Times: []time.Time{tm, tm2},
	}, "times", "2006-01-05T15:04:05Z")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Times: []time.Time{tm, tm2},
	}, "times", "2006-01-10T15:04:05Z")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Times: []time.Time{tm, tm2},
	}, "times", "2006-01-11T15:04:05Z")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualDate(t *testing.T) {
	t.Parallel()

	d, err := civil.ParseDate("2006-01-02")
	require.Nil(t, err)

	match, err := GreaterEqual(&testType1{
		Date: d,
	}, "date", "2006-01-01")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Date: d,
	}, "date", "2006-01-02")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Date: d,
	}, "date", "2006-01-03")
	require.Nil(t, err)
	require.False(t, match)
}

func TestGreaterEqualDates(t *testing.T) {
	t.Parallel()

	d1, err := civil.ParseDate("2006-01-01")
	require.Nil(t, err)

	d2, err := civil.ParseDate("2006-01-03")
	require.Nil(t, err)

	match, err := GreaterEqual(&testType1{
		Dates: []civil.Date{d1, d2},
	}, "dates", "2006-01-02")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Dates: []civil.Date{d1, d2},
	}, "dates", "2006-01-03")
	require.Nil(t, err)
	require.True(t, match)

	match, err = GreaterEqual(&testType1{
		Dates: []civil.Date{d1, d2},
	}, "dates", "2006-01-04")
	require.Nil(t, err)
	require.False(t, match)
}
