package path_test

import (
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/firestuff/patchy/path"
	"github.com/stretchr/testify/require"
)

type testType1 struct {
	Int     int
	Int64   int64
	UInt    uint
	UInt64  uint64
	Float32 float32
	Float64 float64
	String  string `json:"string2,omitempty"`
	Bool    bool   `json:"bool2"`
	BoolP   *bool

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
	Date  civil.Date
	Dates []civil.Date

	TimeP  *time.Time
	TimesP []*time.Time
}

type testType2 struct {
	Tt1  testType1
	Tt1p *testType1
}

func TestSet(t *testing.T) {
	t.Parallel()

	tt1 := &testType1{}
	err := path.Set(tt1, "int64", "1234")
	require.Nil(t, err)
	require.Equal(t, int64(1234), tt1.Int64)

	err = path.Set(tt1, "time", "2022-11-01-08:00")
	require.Nil(t, err)
	require.Equal(t, int64(1667289600), tt1.Time.Unix())

	tt2 := &testType2{}
	err = path.Set(tt2, "tt1p.bool2", "true")
	require.Nil(t, err)
	require.Equal(t, true, tt2.Tt1p.Bool)

	err = path.Set(tt2, "tt1p.string2", "foo")
	require.Nil(t, err)
	require.Equal(t, "foo", tt2.Tt1p.String)

	err = path.Set(tt2, "tt1.boolp", "true")
	require.Nil(t, err)
	require.Equal(t, true, *tt2.Tt1.BoolP)
}
