package storebus

import "os"
import "testing"

import "github.com/firestuff/patchy/metadata"
import "github.com/firestuff/patchy/store"

func TestStoreBus(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	sb := NewStoreBus(store.NewLocalStore(dir))

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

	if out1.ETag != "11cb2d0f4dddf836245d5cc0b667e1275b3c0e10777b29335985cfd97210bbbb" {
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

	if out3.ETag != "efce5d60be6fd043869c0dde09ac3477f1687fc36118ba68d82114b45549a800" {
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

	sb := NewStoreBus(store.NewLocalStore(dir))

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
