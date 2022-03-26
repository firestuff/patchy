package storebus

import "os"
import "testing"

import "github.com/firestuff/patchy/metadata"

func TestStoreBus(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	sb := NewStoreBus(dir)

	err = sb.Write("storeBusTest", &storeBusTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
		Opaque: "foo",
	})
	if err != nil {
		t.Fatal(err)
	}

	out1 := &storeBusTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	}

	err = sb.Read("storeBusTest", out1)
	if err != nil {
		t.Fatal(err)
	}

	if out1.Opaque != "foo" {
		t.Errorf("%+v", out1)
	}

	if out1.Sha256 != "5d6f98d2ff3b70bcd32c4ac16625c20456c97d8f16c7cbb21c36514268933ec5" {
		t.Errorf("%+v", out1)
	}

	ch := sb.Subscribe("storeBusTest", &storeBusTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	})

	sb.Write("storeBusTest", &storeBusTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
		Opaque: "bar",
	})

	out3 := (<-ch).(*storeBusTest)

	if out3.Opaque != "bar" {
		t.Errorf("%+v", out3)
	}

	if out3.Sha256 != "e17f86cb37e9af977b4345fa542096d8974237bdad2c59e0be5d3975dffdc42e" {
		t.Errorf("%+v", out3)
	}
}

func TestStoreBusDelete(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	sb := NewStoreBus(dir)

	ch := sb.Subscribe("storeBusTest", &storeBusTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	})

	sb.Write("storeBusTest", &storeBusTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
		Opaque: "foo",
	})

	out := (<-ch).(*storeBusTest)
	if out.Opaque != "foo" {
		t.Errorf("%+v", out)
	}

	sb.Delete("storeBusTest", &storeBusTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	})

	out2, ok := <-ch
	if ok {
		t.Errorf("%+v", out2)
	}
}

type storeBusTest struct {
	metadata.Metadata
	Opaque string
}
