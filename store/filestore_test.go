package store_test

import (
	"os"
	"testing"

	"github.com/firestuff/patchy/metadata"
	"github.com/firestuff/patchy/store"
	"github.com/stretchr/testify/require"
)

func TestFileStore(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	defer os.RemoveAll(dir)

	store := store.NewFileStore(dir)

	err = store.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			ID: "id1",
		},
		Opaque: "foo",
	})
	require.Nil(t, err)

	err = store.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			ID: "id2",
		},
		Opaque: "bar",
	})
	require.Nil(t, err)

	out1, err := store.Read("storeTest", "id1", newStoreTest)
	require.Nil(t, err)
	require.Equal(t, "foo", out1.(*storeTest).Opaque)

	out2, err := store.Read("storeTest", "id2", newStoreTest)
	require.Nil(t, err)
	require.Equal(t, "bar", out2.(*storeTest).Opaque)
}

func TestFileStoreDelete(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	defer os.RemoveAll(dir)

	store := store.NewFileStore(dir)

	err = store.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			ID: "id1",
		},
		Opaque: "foo",
	})
	require.Nil(t, err)

	out1, err := store.Read("storeTest", "id1", newStoreTest)
	require.Nil(t, err)
	require.Equal(t, "foo", out1.(*storeTest).Opaque)

	err = store.Delete("storeTest", "id1")
	require.Nil(t, err)

	out2, err := store.Read("storeTest", "id1", newStoreTest)
	require.Nil(t, err)
	require.Nil(t, out2)
}

func TestFileStoreList(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	defer os.RemoveAll(dir)

	store := store.NewFileStore(dir)

	objs, err := store.List("storeTest", func() any { return &storeTest{} })
	require.Nil(t, err)
	require.Len(t, objs, 0)

	err = store.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			ID: "id1",
		},
		Opaque: "foo",
	})
	require.Nil(t, err)

	err = store.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			ID: "id2",
		},
		Opaque: "bar",
	})
	require.Nil(t, err)

	objs, err = store.List("storeTest", func() any { return &storeTest{} })
	require.Nil(t, err)
	require.Len(t, objs, 2)
	require.ElementsMatch(t, []string{"foo", "bar"}, []string{objs[0].(*storeTest).Opaque, objs[1].(*storeTest).Opaque})
}

type storeTest struct {
	metadata.Metadata
	Opaque string
}

func newStoreTest() any {
	return &storeTest{}
}
