package storebus_test

import (
	"os"
	"testing"

	"github.com/firestuff/patchy/metadata"
	"github.com/firestuff/patchy/store"
	"github.com/firestuff/patchy/storebus"
	"github.com/stretchr/testify/require"
)

func TestStoreBus(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	defer os.RemoveAll(dir)

	sb := storebus.NewStoreBus(store.NewFileStore(dir))

	err = sb.Write("storeBusTest", &storeBusTest{
		Metadata: metadata.Metadata{
			ID: "id1",
		},
		Opaque: "foo",
	})
	require.Nil(t, err)

	c1, err := sb.ReadStream("storeBusTest", "id1", newStoreBusTest)
	require.Nil(t, err)

	defer sb.CloseReadStream("storeBusTest", "id1", c1)

	c2, err := sb.ListStream("storeBusTest", newStoreBusTest)
	require.Nil(t, err)

	defer sb.CloseListStream("storeBusTest", c2)

	out1 := (<-c1).(*storeBusTest)
	require.Equal(t, "foo", out1.Opaque)
	require.Equal(t, "etag:2c8edc6414452b8dee7826bd55e585f850ac47a0dcfc357dc1fcaaa3164cdfa2", out1.ETag)

	l1 := <-c2
	require.Len(t, l1, 1)
	require.Equal(t, "foo", l1[0].(*storeBusTest).Opaque)
	require.Equal(t, "etag:2c8edc6414452b8dee7826bd55e585f850ac47a0dcfc357dc1fcaaa3164cdfa2", l1[0].(*storeBusTest).ETag)

	err = sb.Write("storeBusTest", &storeBusTest{
		Metadata: metadata.Metadata{
			ID: "id1",
		},
		Opaque: "bar",
	})
	require.Nil(t, err)

	out2 := (<-c1).(*storeBusTest)
	require.Equal(t, "bar", out2.Opaque)
	require.Equal(t, "etag:906fda69e9893280ca9294bd04eb276794da9a8904fc0b671c69175f08cc03c6", out2.ETag)

	l2 := <-c2
	require.Len(t, l2, 1)
	require.Equal(t, "bar", l2[0].(*storeBusTest).Opaque)
	require.Equal(t, "etag:906fda69e9893280ca9294bd04eb276794da9a8904fc0b671c69175f08cc03c6", l2[0].(*storeBusTest).ETag)

	l2a, err := sb.List("storeBusTest", newStoreBusTest)
	require.Nil(t, err)
	require.Len(t, l2a, 1)
	require.Equal(t, "bar", l2a[0].(*storeBusTest).Opaque)
	require.Equal(t, "etag:906fda69e9893280ca9294bd04eb276794da9a8904fc0b671c69175f08cc03c6", l2a[0].(*storeBusTest).ETag)

	out2a, err := sb.Read("storeBusTest", "id1", newStoreBusTest)
	require.Nil(t, err)
	require.Equal(t, "bar", out2a.(*storeBusTest).Opaque)
	require.Equal(t, "etag:906fda69e9893280ca9294bd04eb276794da9a8904fc0b671c69175f08cc03c6", out2a.(*storeBusTest).ETag)

	err = sb.Write("storeBusTest", &storeBusTest{
		Metadata: metadata.Metadata{
			ID: "id2",
		},
		Opaque: "zig",
	})
	require.Nil(t, err)

	l3 := <-c2
	require.Len(t, l3, 2)
	require.ElementsMatch(
		t,
		[]string{
			l3[0].(*storeBusTest).Opaque,
			l3[1].(*storeBusTest).Opaque,
		},
		[]string{
			"bar",
			"zig",
		},
	)

	l3a, err := sb.List("storeBusTest", newStoreBusTest)
	require.Nil(t, err)
	require.Len(t, l3a, 2)
	require.ElementsMatch(
		t,
		[]string{
			l3a[0].(*storeBusTest).Opaque,
			l3a[1].(*storeBusTest).Opaque,
		},
		[]string{
			"bar",
			"zig",
		},
	)
}

func TestStoreBusDelete(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	defer os.RemoveAll(dir)

	sb := storebus.NewStoreBus(store.NewFileStore(dir))

	c1, err := sb.ReadStream("storeBusTest", "id1", newStoreBusTest)
	require.Nil(t, err)

	defer sb.CloseReadStream("storeBusTest", "id1", c1)

	preout := <-c1
	require.Nil(t, preout)

	err = sb.Write("storeBusTest", &storeBusTest{
		Metadata: metadata.Metadata{
			ID: "id1",
		},
		Opaque: "foo",
	})
	require.Nil(t, err)

	out := (<-c1).(*storeBusTest)
	require.Equal(t, "foo", out.Opaque)

	err = sb.Delete("storeBusTest", "id1")
	require.Nil(t, err)

	_, ok := <-c1
	require.False(t, ok)
}

type storeBusTest struct {
	metadata.Metadata
	Opaque string
}

func newStoreBusTest() any {
	return &storeBusTest{}
}
