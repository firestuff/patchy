package path

import "testing"
import "time"

import "cloud.google.com/go/civil"
import "github.com/stretchr/testify/require"

func TestEqualStruct(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType2{
		Tt1: testType1{
			Int: 2345,
		},
	}, "tt1.int", "2345")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType2{
		Tt1p: &testType1{
			Int: 2345,
		},
	}, "tt1p.int", "2345")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType2{}, "tt1p.int", "2345")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualPointer(t *testing.T) {
	t.Parallel()

	tm, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	require.Nil(t, err)

	match, err := Equal(&testType1{
		TimeP: &tm,
	}, "timep", "2006-01-02T15:04:05Z")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		TimeP: &tm,
	}, "timep", "2006-01-02T15:04:05+01:00")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualPointers(t *testing.T) {
	t.Parallel()

	tm1, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	require.Nil(t, err)

	tm2, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-10T15:04:05Z")
	require.Nil(t, err)

	match, err := Equal(&testType1{
		TimesP: []*time.Time{&tm1, nil, &tm2},
	}, "timesp", "2006-01-10T15:04:05Z")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		TimesP: []*time.Time{&tm1, &tm2},
	}, "timesp", "2006-01-02T15:04:05+01:00")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualInt(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType1{
		Int: 1234,
	}, "int", "1234")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Int: 1234,
	}, "int", "1235")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualInt64(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType1{
		Int64: 3456,
	}, "int64", "3456")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Int64: 3456,
	}, "int64", "3457")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualUInt(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType1{
		UInt: 4567,
	}, "uint", "4567")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		UInt: 4567,
	}, "uint", "4568")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualUInt64(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType1{
		UInt64: 5678,
	}, "uint64", "5678")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		UInt64: 5678,
	}, "uint64", "5679")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualFloat32(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType1{
		Float32: 3.1415,
	}, "float32", "3.1415")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Float32: 3.1415,
	}, "float32", "3.1416")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualFloat64(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType1{
		Float64: 3.14159265,
	}, "float64", "3.14159265")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Float64: 3.14159265,
	}, "float64", "3.14159266")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualString(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType1{
		String: "foo",
	}, "string", "foo")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		String: "foo",
	}, "string", "bar")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualBool(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType1{
		Bool: true,
	}, "bool", "true")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Bool: true,
	}, "bool", "false")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualInts(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType1{
		Ints: []int{2, 4, 7},
	}, "ints", "4")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Ints: []int{2, 4, 7},
	}, "ints", "5")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualInt64s(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType1{
		Int64s: []int64{2, 4, 7},
	}, "int64s", "4")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Int64s: []int64{2, 4, 7},
	}, "int64s", "5")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualUInts(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType1{
		UInts: []uint{2, 4, 7},
	}, "uints", "4")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		UInts: []uint{2, 4, 7},
	}, "uints", "5")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualUInt64s(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType1{
		UInt64s: []uint64{2, 4, 7},
	}, "uint64s", "4")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		UInt64s: []uint64{2, 4, 7},
	}, "uint64s", "5")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualFloat32s(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType1{
		Float32s: []float32{3.1415, 2.7182},
	}, "float32s", "2.7182")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Float32s: []float32{3.1415, 2.7182},
	}, "float32s", "2.7183")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualFloat64s(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType1{
		Float64s: []float64{3.1415, 2.7182},
	}, "float64s", "2.7182")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Float64s: []float64{3.1415, 2.7182},
	}, "float64s", "2.7183")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualStrings(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType1{
		Strings: []string{"foo", "bar"},
	}, "strings", "foo")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Strings: []string{"foo", "bar"},
	}, "strings", "zig")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualBools(t *testing.T) {
	t.Parallel()

	match, err := Equal(&testType1{
		Bools: []bool{true, false},
	}, "bools", "true")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Bools: []bool{false, false},
	}, "bools", "true")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualTime(t *testing.T) {
	t.Parallel()

	tm, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	require.Nil(t, err)

	match, err := Equal(&testType1{
		Time: tm,
	}, "time", "2006-01-02T15:04:05Z")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Time: tm,
	}, "time", "2006-01-02T15:04:05+00:00")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Time: tm,
	}, "time", "2006-01-02T15:04:05+01:00")
	require.Nil(t, err)
	require.False(t, match)

	match, err = Equal(&testType1{
		Time: tm,
	}, "time", "1136214245")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Time: tm,
	}, "time", "1136214246")
	require.Nil(t, err)
	require.False(t, match)

	match, err = Equal(&testType1{
		Time: tm,
	}, "time", "1136214245000")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Time: tm,
	}, "time", "1136214245001")
	require.Nil(t, err)
	require.False(t, match)

	tm2, err := time.Parse("2006-01-02T15:04:05.999999999Z", "2006-01-02T15:04:05.500000000Z")
	require.Nil(t, err)

	match, err = Equal(&testType1{
		Time: tm2,
	}, "time", "1136214245")
	require.Nil(t, err)
	require.True(t, match)
}

func TestEqualTimes(t *testing.T) {
	t.Parallel()

	tm, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	require.Nil(t, err)

	tm2, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-10T15:04:05Z")
	require.Nil(t, err)

	match, err := Equal(&testType1{
		Times: []time.Time{tm, tm2},
	}, "times", "1136214245000")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Times: []time.Time{tm, tm2},
	}, "times", "1136214245001")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualDate(t *testing.T) {
	t.Parallel()

	d, err := civil.ParseDate("2006-01-01")
	require.Nil(t, err)

	match, err := Equal(&testType1{
		Date: d,
	}, "date", "2006-01-01")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Date: d,
	}, "date", "2006-01-02")
	require.Nil(t, err)
	require.False(t, match)
}

func TestEqualDates(t *testing.T) {
	t.Parallel()

	d1, err := civil.ParseDate("2006-01-01")
	require.Nil(t, err)

	d2, err := civil.ParseDate("2006-01-02")
	require.Nil(t, err)

	match, err := Equal(&testType1{
		Dates: []civil.Date{d1, d2},
	}, "dates", "2006-01-02")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Equal(&testType1{
		Dates: []civil.Date{d1, d2},
	}, "dates", "2006-01-03")
	require.Nil(t, err)
	require.False(t, match)
}
