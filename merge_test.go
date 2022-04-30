package patchy_test

import (
	"testing"

	"github.com/firestuff/patchy"
	"github.com/stretchr/testify/require"
)

type MergeTestType struct {
	A string
	B int64
	C []string
	D NestedType
	E *NestedType
}

type NestedType struct {
	F []int64
}

func TestMerge(t *testing.T) {
	t.Parallel()

	to := &MergeTestType{
		A: "foo",
		B: 42,
		C: []string{"zig", "zag"},
		D: NestedType{
			F: []int64{42, 43},
		},
		E: &NestedType{
			F: []int64{44, 45},
		},
	}

	err := patchy.Merge(to, &MergeTestType{
		A: "bar",
	})
	require.Nil(t, err)
	require.Equal(t, "bar", to.A)
	require.Equal(t, int64(42), to.B)

	err = patchy.Merge(to, &MergeTestType{
		B: 46,
		C: []string{"ooh", "aah"},
		D: NestedType{
			F: []int64{47, 48},
		},
		E: &NestedType{
			F: []int64{49, 50},
		},
	})
	require.Nil(t, err)
	require.Equal(t, "bar", to.A)
	require.Equal(t, int64(46), to.B)
	require.Equal(t, []string{"ooh", "aah"}, to.C)
	require.Equal(t, []int64{47, 48}, to.D.F)
	require.Equal(t, []int64{49, 50}, to.E.F)
}
