package store

import "os"
import "testing"

import "github.com/firestuff/patchy/metadata"

func TestStore(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store := NewStore(dir)

	err = store.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
		Opaque: "foo",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = store.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			Id: "id2",
		},
		Opaque: "bar",
	})
	if err != nil {
		t.Fatal(err)
	}

	out1 := &storeTest{
		Metadata: metadata.Metadata{
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
		Metadata: metadata.Metadata{
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

func TestDelete(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store := NewStore(dir)

	err = store.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
		Opaque: "foo",
	})
	if err != nil {
		t.Fatal(err)
	}

	out1 := &storeTest{
		Metadata: metadata.Metadata{
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
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = store.Read("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	})
	if err == nil {
		t.Fatal("Read() succeeded")
	}
}

func TestList(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store := NewStore(dir)

	err = store.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
		Opaque: "foo",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = store.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			Id: "id2",
		},
		Opaque: "bar",
	})
	if err != nil {
		t.Fatal(err)
	}

	objs, err := store.List("storeTest", func() any { return &storeTest{} })
	if err != nil {
		t.Fatal(err)
	}

	if len(objs) != 2 {
		t.Fatalf("%+v", objs)
	}

	if !((objs[0].(*storeTest).Opaque == "foo" && objs[1].(*storeTest).Opaque == "bar") ||
		(objs[0].(*storeTest).Opaque == "bar" && objs[1].(*storeTest).Opaque == "foo")) {
		t.Fatalf("%+v", objs)
	}
}

type storeTest struct {
	metadata.Metadata
	Opaque string
}
