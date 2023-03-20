//nolint:testpackage
package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type mergeTestType struct {
	A string
	B int
	C []string
	D nestedType
	E *nestedType
}

type nestedType struct {
	F []int
}

func TestMergeString(t *testing.T) {
	t.Parallel()

	to := &mergeTestType{
		A: "foo",
		B: 42,
	}

	merge(to, &mergeTestType{
		A: "bar",
	})

	require.Equal(t, "bar", to.A)
	require.Equal(t, 42, to.B)
}

func TestMergeSlice(t *testing.T) {
	t.Parallel()

	to := &mergeTestType{
		B: 42,
		C: []string{"foo", "bar"},
	}

	merge(to, &mergeTestType{
		C: []string{"zig", "zag"},
	})

	require.Equal(t, 42, to.B)
	require.Equal(t, []string{"zig", "zag"}, to.C)
}

func TestMergeNested(t *testing.T) {
	t.Parallel()

	to := &mergeTestType{
		B: 42,
		D: nestedType{
			F: []int{42, 43},
		},
	}

	merge(to, &mergeTestType{
		D: nestedType{
			F: []int{44, 45},
		},
	})

	require.Equal(t, 42, to.B)
	require.Equal(t, []int{44, 45}, to.D.F)
}

func TestMergeNestedPointer(t *testing.T) {
	t.Parallel()

	to := &mergeTestType{
		B: 42,
		E: &nestedType{
			F: []int{42, 43},
		},
	}

	merge(to, &mergeTestType{
		E: &nestedType{
			F: []int{49, 50},
		},
	})

	require.Equal(t, 42, to.B)
	require.Equal(t, []int{49, 50}, to.E.F)
}
