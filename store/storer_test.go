package store_test

import (
	"testing"

	"github.com/firestuff/patchy/metadata"
	"github.com/firestuff/patchy/store"
	"github.com/stretchr/testify/require"
)

func testStorer(t *testing.T, st store.Storer) {
	err := st.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			ID: "id1",
		},
		Opaque: "foo",
	})
	require.Nil(t, err)

	err = st.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			ID: "id2",
		},
		Opaque: "bar",
	})
	require.Nil(t, err)

	err = st.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			ID: "id2",
		},
		Opaque: "zig",
	})
	require.Nil(t, err)

	out1, err := st.Read("storeTest", "id1", newStoreTest)
	require.Nil(t, err)
	require.NotNil(t, out1)
	require.Equal(t, "foo", out1.(*storeTest).Opaque)

	out2, err := st.Read("storeTest", "id2", newStoreTest)
	require.Nil(t, err)
	require.NotNil(t, out1)
	require.Equal(t, "zig", out2.(*storeTest).Opaque)
}

func testDelete(t *testing.T, st store.Storer) {
	err := st.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			ID: "id1",
		},
		Opaque: "foo",
	})
	require.Nil(t, err)

	out1, err := st.Read("storeTest", "id1", newStoreTest)
	require.Nil(t, err)
	require.Equal(t, "foo", out1.(*storeTest).Opaque)

	err = st.Delete("storeTest", "id1")
	require.Nil(t, err)

	out2, err := st.Read("storeTest", "id1", newStoreTest)
	require.Nil(t, err)
	require.Nil(t, out2)
}

func testList(t *testing.T, st store.Storer) {
	objs, err := st.List("storeTest", func() any { return &storeTest{} })
	require.Nil(t, err)
	require.Len(t, objs, 0)

	err = st.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			ID: "id1",
		},
		Opaque: "foo",
	})
	require.Nil(t, err)

	err = st.Write("storeTest", &storeTest{
		Metadata: metadata.Metadata{
			ID: "id2",
		},
		Opaque: "bar",
	})
	require.Nil(t, err)

	objs, err = st.List("storeTest", func() any { return &storeTest{} })
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