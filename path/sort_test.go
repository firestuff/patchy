package path

import "testing"
import "time"

import "github.com/stretchr/testify/require"

func TestSortStruct(t *testing.T) {
	t.Parallel()

	objs := []*testType2{
		&testType2{
			Tt1: testType1{
				Int: 2,
			},
		},
		&testType2{
			Tt1: testType1{
				Int: 1,
			},
		},
		&testType2{
			Tt1: testType1{
				Int: 3,
			},
		},
	}

	err := Sort(objs, "tt1.int")
	require.Nil(t, err)
	require.Len(t, objs, 3)
	require.Equal(t, []int{1, 2, 3}, []int{objs[0].Tt1.Int, objs[1].Tt1.Int, objs[2].Tt1.Int})
}

func TestSortReverse(t *testing.T) {
	t.Parallel()

	objs := []*testType1{
		&testType1{
			Int: 3,
		},
		&testType1{
			Int: 1,
		},
		&testType1{
			Int: 2,
		},
	}

	err := SortReverse(objs, "int")
	require.Nil(t, err)
	require.Len(t, objs, 3)
	require.Equal(t, []int{3, 2, 1}, []int{objs[0].Int, objs[1].Int, objs[2].Int})
}

func TestSortInt(t *testing.T) {
	t.Parallel()

	objs := []*testType1{
		&testType1{
			Int: 3,
		},
		&testType1{
			Int: 1,
		},
		&testType1{
			Int: 2,
		},
	}

	err := Sort(objs, "int")
	require.Nil(t, err)
	require.Len(t, objs, 3)
	require.Equal(t, []int{1, 2, 3}, []int{objs[0].Int, objs[1].Int, objs[2].Int})
}

func TestSortInt64(t *testing.T) {
	t.Parallel()

	objs := []*testType1{
		&testType1{
			Int64: 3,
		},
		&testType1{
			Int64: 1,
		},
		&testType1{
			Int64: 2,
		},
	}

	err := Sort(objs, "int64")
	require.Nil(t, err)
	require.Len(t, objs, 3)
	require.Equal(t, []int64{1, 2, 3}, []int64{objs[0].Int64, objs[1].Int64, objs[2].Int64})
}

func TestSortUint(t *testing.T) {
	t.Parallel()

	objs := []*testType1{
		&testType1{
			UInt: 3,
		},
		&testType1{
			UInt: 1,
		},
		&testType1{
			UInt: 2,
		},
	}

	err := Sort(objs, "uint")
	require.Nil(t, err)
	require.Len(t, objs, 3)
	require.Equal(t, []uint{1, 2, 3}, []uint{objs[0].UInt, objs[1].UInt, objs[2].UInt})
}

func TestSortUint64(t *testing.T) {
	t.Parallel()

	objs := []*testType1{
		&testType1{
			UInt64: 3,
		},
		&testType1{
			UInt64: 1,
		},
		&testType1{
			UInt64: 2,
		},
	}

	err := Sort(objs, "uint64")
	require.Nil(t, err)
	require.Len(t, objs, 3)
	require.Equal(t, []uint64{1, 2, 3}, []uint64{objs[0].UInt64, objs[1].UInt64, objs[2].UInt64})
}

func TestSortFloat32(t *testing.T) {
	t.Parallel()

	objs := []*testType1{
		&testType1{
			Float32: 3.3,
		},
		&testType1{
			Float32: 1.1,
		},
		&testType1{
			Float32: 2.2,
		},
	}

	err := Sort(objs, "float32")
	require.Nil(t, err)
	require.Len(t, objs, 3)
	require.Equal(t, []float32{1.1, 2.2, 3.3}, []float32{objs[0].Float32, objs[1].Float32, objs[2].Float32})
}

func TestSortFloat64(t *testing.T) {
	t.Parallel()

	objs := []*testType1{
		&testType1{
			Float64: 3.3,
		},
		&testType1{
			Float64: 1.1,
		},
		&testType1{
			Float64: 2.2,
		},
	}

	err := Sort(objs, "float64")
	require.Nil(t, err)
	require.Len(t, objs, 3)
	require.Equal(t, []float64{1.1, 2.2, 3.3}, []float64{objs[0].Float64, objs[1].Float64, objs[2].Float64})
}

func TestSortString(t *testing.T) {
	t.Parallel()

	objs := []*testType1{
		&testType1{
			String: "zig",
		},
		&testType1{
			String: "bar",
		},
		&testType1{
			String: "foo",
		},
	}

	err := Sort(objs, "string")
	require.Nil(t, err)
	require.Len(t, objs, 3)
	require.Equal(t, []string{"bar", "foo", "zig"}, []string{objs[0].String, objs[1].String, objs[2].String})
}

func TestSortBool(t *testing.T) {
	t.Parallel()

	objs := []*testType1{
		&testType1{
			Bool: true,
		},
		&testType1{
			Bool: false,
		},
		&testType1{
			Bool: true,
		},
	}

	err := Sort(objs, "bool")
	require.Nil(t, err)
	require.Len(t, objs, 3)
	require.Equal(t, []bool{false, true, true}, []bool{objs[0].Bool, objs[1].Bool, objs[2].Bool})
}

func TestSortTime(t *testing.T) {
	t.Parallel()

	t1, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-01T15:04:05Z")
	require.Nil(t, err)
	t2, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")
	require.Nil(t, err)
	t3, err := time.Parse("2006-01-02T15:04:05Z", "2006-01-03T15:04:05Z")
	require.Nil(t, err)

	objs := []*testType1{
		&testType1{
			Time: t3,
		},
		&testType1{
			Time: t1,
		},
		&testType1{
			Time: t2,
		},
	}

	err = Sort(objs, "time")
	require.Nil(t, err)
	require.Len(t, objs, 3)
	require.Equal(t, []time.Time{t1, t2, t3}, []time.Time{objs[0].Time, objs[1].Time, objs[2].Time})
}
