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
			Id: "id1",
		},
		Opaque: "foo",
	})
	require.Nil(t, err)

	out1 := &storeBusTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	}

	err = sb.Read("storeBusTest", out1)
	require.Nil(t, err)
	require.Equal(t, "foo", out1.Opaque)
	require.Equal(t, "etag:2c8edc6414452b8dee7826bd55e585f850ac47a0dcfc357dc1fcaaa3164cdfa2", out1.ETag)

	ch := sb.SubscribeKey("storeBusTest", "id1")

	err = sb.Write("storeBusTest", &storeBusTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
		Opaque: "bar",
	})
	require.Nil(t, err)

	out3 := (<-ch).(*storeBusTest)
	require.Equal(t, "bar", out3.Opaque)
	require.Equal(t, "etag:906fda69e9893280ca9294bd04eb276794da9a8904fc0b671c69175f08cc03c6", out3.ETag)
}

func TestStoreBusDelete(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	sb := storebus.NewStoreBus(store.NewFileStore(dir))

	ch := sb.SubscribeKey("storeBusTest", "id1")

	err = sb.Write("storeBusTest", &storeBusTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
		Opaque: "foo",
	})
	require.Nil(t, err)

	out := (<-ch).(*storeBusTest)
	require.Equal(t, "foo", out.Opaque)

	err = sb.Delete("storeBusTest", "id1")
	require.Nil(t, err)

	_, ok := <-ch
	require.False(t, ok)
}

type storeBusTest struct {
	metadata.Metadata
	Opaque string
}
