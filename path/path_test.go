package path

import "time"

import "cloud.google.com/go/civil"

type testType1 struct {
	Int     int
	Int64   int64
	UInt    uint
	UInt64  uint64
	Float32 float32
	Float64 float64
	String  string `json:"string2,omitempty"`
	Bool    bool   `json:"bool2"`

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
