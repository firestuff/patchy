package storebus

import "os"
import "testing"

import "github.com/stretchr/testify/require"

import "github.com/firestuff/patchy/metadata"
import "github.com/firestuff/patchy/store"

func TestStoreBus(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	sb := NewStoreBus(store.NewLocalStore(dir))

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
	require.Equal(t, "11cb2d0f4dddf836245d5cc0b667e1275b3c0e10777b29335985cfd97210bbbb", out1.ETag)

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
	require.Equal(t, "bar", out3.Opaque)
	require.Equal(t, "efce5d60be6fd043869c0dde09ac3477f1687fc36118ba68d82114b45549a800", out3.ETag)
}

func TestStoreBusDelete(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)
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
	require.Equal(t, "foo", out.Opaque)

	sb.Delete("storeBusTest", "id1")

	_, ok := <-ch
	require.False(t, ok)
}

type storeBusTest struct {
	metadata.Metadata
	Opaque string
}
