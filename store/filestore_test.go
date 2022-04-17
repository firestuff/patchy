package store

import "os"
import "testing"

import "github.com/stretchr/testify/require"

import "github.com/firestuff/patchy/metadata"

func TestFileStore(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	store := NewFileStore(dir)

	err = store.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
		Opaque: "foo",
	})
	require.Nil(t, err)

	err = store.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			Id: "id2",
		},
		Opaque: "bar",
	})
	require.Nil(t, err)

	out1 := &storeTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	}

	err = store.Read("storeTest", out1)
	require.Nil(t, err)
	require.Equal(t, "foo", out1.Opaque)

	out2 := &storeTest{
		Metadata: metadata.Metadata{
			Id: "id2",
		},
	}

	err = store.Read("storeTest", out2)
	require.Nil(t, err)
	require.Equal(t, "bar", out2.Opaque)
}

func TestFileStoreDelete(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	store := NewFileStore(dir)

	err = store.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
		Opaque: "foo",
	})
	require.Nil(t, err)

	out1 := &storeTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	}

	err = store.Read("storeTest", out1)
	require.Nil(t, err)
	require.Equal(t, "foo", out1.Opaque)

	err = store.Delete("storeTest", "id1")
	require.Nil(t, err)

	err = store.Read("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	})
	require.NotNil(t, err)
}

func TestFileStoreList(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	store := NewFileStore(dir)

	err = store.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
		Opaque: "foo",
	})
	require.Nil(t, err)

	err = store.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			Id: "id2",
		},
		Opaque: "bar",
	})
	require.Nil(t, err)

	objs, err := store.List("storeTest", func() any { return &storeTest{} })
	require.Nil(t, err)
	require.Len(t, objs, 2)
	require.ElementsMatch(t, []string{"foo", "bar"}, []string{objs[0].(*storeTest).Opaque, objs[1].(*storeTest).Opaque})
}

type storeTest struct {
	metadata.Metadata
	Opaque string
}