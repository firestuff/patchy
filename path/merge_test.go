package path_test

import (
	"encoding/json"
	"testing"

	"github.com/firestuff/patchy/path"
	"github.com/stretchr/testify/require"
)

type mergeTestType struct {
	A string
	B int
	C []string
	D nestedType
	E *nestedType
	H int
}

type nestedType struct {
	F []int
	G string
}

func TestMergeString(t *testing.T) {
	t.Parallel()

	to := &mergeTestType{
		A: "foo",
		B: 42,
	}

	path.Merge(to, &mergeTestType{
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

	path.Merge(to, &mergeTestType{
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
			G: "bar",
		},
	}

	path.Merge(to, &mergeTestType{
		D: nestedType{
			F: []int{44, 45},
		},
	})

	require.Equal(t, 42, to.B)
	require.Equal(t, []int{44, 45}, to.D.F)
	require.Equal(t, "bar", to.D.G)
}

func TestMergeNestedPointer(t *testing.T) {
	t.Parallel()

	to := &mergeTestType{
		B: 42,
		E: &nestedType{
			F: []int{42, 43},
			G: "bar",
		},
	}

	path.Merge(to, &mergeTestType{
		E: &nestedType{
			F: []int{49, 50},
		},
	})

	require.Equal(t, 42, to.B)
	require.Equal(t, []int{49, 50}, to.E.F)
	require.Equal(t, "bar", to.E.G)
}

func TestMergeMap(t *testing.T) {
	t.Parallel()

	to := &mergeTestType{
		A: "foo",
		B: 42,
		D: nestedType{
			F: []int{42, 43},
			G: "bar",
		},
		H: 5,
	}

	from := map[string]any{}

	err := json.Unmarshal(
		[]byte(`
{
	"B": 45,
	"D": {
		"F": [46, 47]
	},
	"H": 0
}`),
		&from,
	)
	require.NoError(t, err)

	err = path.MergeMap(to, from)
	require.NoError(t, err)

	require.Equal(t, "foo", to.A)
	require.Equal(t, 45, to.B)
	require.Equal(t, []int{46, 47}, to.D.F)
	require.Equal(t, "bar", to.D.G)
	require.Equal(t, 0, to.H)
}
