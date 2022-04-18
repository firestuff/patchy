package path

import "testing"
import "time"

import "github.com/stretchr/testify/require"

func TestMatchStruct(t *testing.T) {
	t.Parallel()

	match, err := Match(&testType2{
		Tt1: testType1{
			Int: 2345,
		},
	}, "tt1.int", "2345")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType2{
		Tt1p: &testType1{
			Int: 2345,
		},
	}, "tt1p.int", "2345")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType2{}, "tt1p.int", "2345")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchInt(t *testing.T) {
	t.Parallel()

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
}

func TestMatchInt64(t *testing.T) {
	t.Parallel()

	match, err := Match(&testType1{
		Int64: 3456,
	}, "int64", "3456")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Int64: 3456,
	}, "int64", "3457")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchUInt(t *testing.T) {
	t.Parallel()

	match, err := Match(&testType1{
		UInt: 4567,
	}, "uint", "4567")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		UInt: 4567,
	}, "uint", "4568")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchUInt64(t *testing.T) {
	t.Parallel()

	match, err := Match(&testType1{
		UInt64: 5678,
	}, "uint64", "5678")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		UInt64: 5678,
	}, "uint64", "5679")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchFloat32(t *testing.T) {
	t.Parallel()

	match, err := Match(&testType1{
		Float32: 3.1415,
	}, "float32", "3.1415")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Float32: 3.1415,
	}, "float32", "3.1416")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchFloat64(t *testing.T) {
	t.Parallel()

	match, err := Match(&testType1{
		Float64: 3.14159265,
	}, "float64", "3.14159265")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Float64: 3.14159265,
	}, "float64", "3.14159266")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchString(t *testing.T) {
	t.Parallel()

	match, err := Match(&testType1{
		String: "foo",
	}, "string", "foo")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		String: "foo",
	}, "string", "bar")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchBool(t *testing.T) {
	t.Parallel()

	match, err := Match(&testType1{
		Bool: true,
	}, "bool", "true")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Bool: true,
	}, "bool", "false")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchInts(t *testing.T) {
	t.Parallel()

	match, err := Match(&testType1{
		Ints: []int{2, 4, 7},
	}, "ints", "4")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Ints: []int{2, 4, 7},
	}, "ints", "5")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchInt64s(t *testing.T) {
	t.Parallel()

	match, err := Match(&testType1{
		Int64s: []int64{2, 4, 7},
	}, "int64s", "4")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Int64s: []int64{2, 4, 7},
	}, "int64s", "5")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchUInts(t *testing.T) {
	t.Parallel()

	match, err := Match(&testType1{
		UInts: []uint{2, 4, 7},
	}, "uints", "4")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		UInts: []uint{2, 4, 7},
	}, "uints", "5")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchUInt64s(t *testing.T) {
	t.Parallel()

	match, err := Match(&testType1{
		UInt64s: []uint64{2, 4, 7},
	}, "uint64s", "4")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		UInt64s: []uint64{2, 4, 7},
	}, "uint64s", "5")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchFloat32s(t *testing.T) {
	t.Parallel()

	match, err := Match(&testType1{
		Float32s: []float32{3.1415, 2.7182},
	}, "float32s", "2.7182")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Float32s: []float32{3.1415, 2.7182},
	}, "float32s", "2.7183")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchFloat64s(t *testing.T) {
	t.Parallel()

	match, err := Match(&testType1{
		Float64s: []float64{3.1415, 2.7182},
	}, "float64s", "2.7182")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Float64s: []float64{3.1415, 2.7182},
	}, "float64s", "2.7183")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchStrings(t *testing.T) {
	t.Parallel()

	match, err := Match(&testType1{
		Strings: []string{"foo", "bar"},
	}, "strings", "foo")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Strings: []string{"foo", "bar"},
	}, "strings", "zig")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchBools(t *testing.T) {
	t.Parallel()

	match, err := Match(&testType1{
		Bools: []bool{true, false},
	}, "bools", "true")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Bools: []bool{false, false},
	}, "bools", "true")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchTime(t *testing.T) {
	t.Parallel()

	tm, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	require.Nil(t, err)

	match, err := Match(&testType1{
		Time: tm,
	}, "time", "2006-01-02T15:04:05Z")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Time: tm,
	}, "time", "2006-01-02T15:04:05+00:00")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Time: tm,
	}, "time", "2006-01-02T15:04:05+01:00")
	require.Nil(t, err)
	require.False(t, match)

	match, err = Match(&testType1{
		Time: tm,
	}, "time", "1136214245")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Time: tm,
	}, "time", "1136214246")
	require.Nil(t, err)
	require.False(t, match)

	match, err = Match(&testType1{
		Time: tm,
	}, "time", "1136214245000")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Time: tm,
	}, "time", "1136214245001")
	require.Nil(t, err)
	require.False(t, match)
}

func TestMatchTimes(t *testing.T) {
	t.Parallel()

	tm, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	require.Nil(t, err)

	tm2, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-10T15:04:05Z")
	require.Nil(t, err)

	match, err := Match(&testType1{
		Times: []time.Time{tm, tm2},
	}, "times", "1136214245000")
	require.Nil(t, err)
	require.True(t, match)

	match, err = Match(&testType1{
		Times: []time.Time{tm, tm2},
	}, "times", "1136214245001")
	require.Nil(t, err)
	require.False(t, match)
}

type testType1 struct {
	Int     int
	Int64   int64
	UInt    uint
	UInt64  uint64
	Float32 float32
	Float64 float64
	String  string
	Bool    bool

	Ints     []int
	Int64s   []int64
	UInts    []uint
	UInt64s  []uint64
	Float32s []float32
	Float64s []float64
	Strings  []string
	Bools    []bool

	Time  time.Time
	Times []time.Time
}

type testType2 struct {
	Tt1  testType1
	Tt1p *testType1
}