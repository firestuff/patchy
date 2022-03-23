package storebus

import "os"
import "testing"

func TestStore(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store := NewStore(dir)

	err = store.Write("storeTest", &storeTest{
		Metadata: Metadata{
			Id: "id1",
		},
		Opaque: "foo",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = store.Write("storeTest", &storeTest{
		Metadata: Metadata{
			Id: "id2",
		},
		Opaque: "bar",
	})
	if err != nil {
		t.Fatal(err)
	}

	out1 := &storeTest{
		Metadata: Metadata{
			Id: "id1",
		},
	}

	err = store.Read("storeTest", out1)
	if err != nil {
		t.Fatal(err)
	}

	if out1.Opaque != "foo" {
		t.Errorf("%+v", out1)
	}

	out2 := &storeTest{
		Metadata: Metadata{
			Id: "id2",
		},
	}

	err = store.Read("storeTest", out2)
	if err != nil {
		t.Fatal(err)
	}

	if out2.Opaque != "bar" {
		t.Errorf("%+v", out2)
	}
}

func TestStoreDelete(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store := NewStore(dir)

	err = store.Write("storeTest", &storeTest{
		Metadata: Metadata{
			Id: "id1",
		},
		Opaque: "foo",
	})
	if err != nil {
		t.Fatal(err)
	}

	out1 := &storeTest{
		Metadata: Metadata{
			Id: "id1",
		},
	}

	err = store.Read("storeTest", out1)
	if err != nil {
		t.Fatal(err)
	}

	if out1.Opaque != "foo" {
		t.Errorf("%+v", out1)
	}

	err = store.Delete("storeTest", &storeTest{
		Metadata: Metadata{
			Id: "id1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = store.Read("storeTest", &storeTest{
		Metadata: Metadata{
			Id: "id1",
		},
	})
	if err == nil {
		t.Fatal("Read() succeeded")
	}
}

type storeTest struct {
	Metadata
	Opaque string
}
