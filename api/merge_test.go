package api

import "testing"

import "golang.org/x/exp/slices"

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

	err := merge(to, &MergeTestType{
		A: "bar",
	})
	if err != nil {
		t.Fatal(err)
	}

	if to.A != "bar" ||
		to.B != 42 {
		t.Fatalf("%+v", to)
	}

	err = merge(to, &MergeTestType{
		B: 46,
		C: []string{"ooh", "aah"},
		D: NestedType{
			F: []int64{47, 48},
		},
		E: &NestedType{
			F: []int64{49, 50},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if to.A != "bar" ||
		to.B != 46 ||
		!slices.Equal(to.C, []string{"ooh", "aah"}) ||
		!slices.Equal(to.D.F, []int64{47, 48}) ||
		!slices.Equal(to.E.F, []int64{49, 50}) {
		t.Fatalf("%+v", to)
	}
}
