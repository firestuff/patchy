package patchy

import "os"
import "testing"

func TestStoreBus(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	sb := NewStoreBus(dir)

	err = sb.Write("storeBusTest", &storeBusTest{
		Metadata: Metadata{
			Id: "id1",
		},
		Opaque: "foo",
	})
	if err != nil {
		t.Fatal(err)
	}

	out1 := &storeBusTest{
		Metadata: Metadata{
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

	ch := sb.Subscribe("storeBusTest", &storeBusTest{
		Metadata: Metadata{
			Id: "id1",
		},
	})

	sb.Write("storeBusTest", &storeBusTest{
		Metadata: Metadata{
			Id: "id1",
		},
		Opaque: "bar",
	})

	out3 := (<-ch).(*storeBusTest)
	if out3.Opaque != "bar" {
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
		Metadata: Metadata{
			Id: "id1",
		},
	})

	sb.Write("storeBusTest", &storeBusTest{
		Metadata: Metadata{
			Id: "id1",
		},
		Opaque: "foo",
	})

	out := (<-ch).(*storeBusTest)
	if out.Opaque != "foo" {
		t.Errorf("%+v", out)
	}

	sb.Delete("storeBusTest", &storeBusTest{
		Metadata: Metadata{
			Id: "id1",
		},
	})

	out2, ok := <-ch
	if ok {
		t.Errorf("%+v", out2)
	}
}

type storeBusTest struct {
	Metadata
	Opaque string
}
