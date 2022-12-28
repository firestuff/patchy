package storebus_test

import (
	"context"
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

	ev1, err := sb.Read(context.Background(), "storeBusTest", "id1", newStoreBusTest)
	require.Nil(t, err)

	ev2, err := sb.List(context.Background(), "storeBusTest", newStoreBusTest)
	require.Nil(t, err)

	out1 := (<-ev1.Chan()).(*storeBusTest)
	require.Equal(t, "foo", out1.Opaque)
	require.Equal(t, "etag:2c8edc6414452b8dee7826bd55e585f850ac47a0dcfc357dc1fcaaa3164cdfa2", out1.ETag)

	l1 := <-ev2.Chan()
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

	out2 := (<-ev1.Chan()).(*storeBusTest)
	require.Equal(t, "bar", out2.Opaque)
	require.Equal(t, "etag:906fda69e9893280ca9294bd04eb276794da9a8904fc0b671c69175f08cc03c6", out2.ETag)

	l2 := <-ev2.Chan()
	require.Len(t, l2, 1)
	require.Equal(t, "bar", l2[0].(*storeBusTest).Opaque)
	require.Equal(t, "etag:906fda69e9893280ca9294bd04eb276794da9a8904fc0b671c69175f08cc03c6", l2[0].(*storeBusTest).ETag)

	err = sb.Write("storeBusTest", &storeBusTest{
		Metadata: metadata.Metadata{
			ID: "id2",
		},
		Opaque: "zig",
	})
	require.Nil(t, err)

	l3 := <-ev2.Chan()
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
}

func TestStoreBusDelete(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	defer os.RemoveAll(dir)

	sb := storebus.NewStoreBus(store.NewFileStore(dir))

	ev1, err := sb.Read(context.Background(), "storeBusTest", "id1", newStoreBusTest)
	require.Nil(t, err)

	preout := <-ev1.Chan()
	require.Nil(t, preout)

	err = sb.Write("storeBusTest", &storeBusTest{
		Metadata: metadata.Metadata{
			ID: "id1",
		},
		Opaque: "foo",
	})
	require.Nil(t, err)

	out := (<-ev1.Chan()).(*storeBusTest)
	require.Equal(t, "foo", out.Opaque)

	err = sb.Delete("storeBusTest", "id1")
	require.Nil(t, err)

	_, ok := <-ev1.Chan()
	require.False(t, ok)
}

type storeBusTest struct {
	metadata.Metadata
	Opaque string
}

func newStoreBusTest() any {
	return &storeBusTest{}
}
