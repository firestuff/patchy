package metadata

import "testing"

type TestType struct {
	Metadata
	Text string
}

func TestMetadata(t *testing.T) {
	t.Parallel()

	tt := &TestType{}

	// Verify promoted field
	tt.Id = "abc123"

	m := GetMetadata(tt)
	if m == nil {
		t.Fatal("GetMetadata")
	}

	if m.Id != "abc123" {
		t.Fatal(m.Id)
	}

	sid := m.GetSafeId()
	if sid != "616263313233" {
		t.Fatal(sid)
	}

	key := m.GetKey("testtype")
	if key != "testtype:616263313233" {
		t.Fatal(key)
	}

	ClearMetadata(tt)

	if tt.Id != "" {
		t.Fatal(tt.Id)
	}
}
