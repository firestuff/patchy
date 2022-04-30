package bus_test

import (
	"testing"

	"github.com/firestuff/patchy/bus"
	"github.com/firestuff/patchy/metadata"
	"github.com/stretchr/testify/require"
)

func TestBus(t *testing.T) {
	t.Parallel()

	bus := bus.NewBus()

	// Announce with no subscribers
	bus.Announce("busTest1", &busTest{
		Metadata: metadata.Metadata{
			ID: "id-nosub",
		},
	})

	// Complex subscription layout
	ch1a := bus.SubscribeKey("busTest1", "id-overlap")
	ch2a := bus.SubscribeKey("busTest2", "id-overlap")
	ch2b := bus.SubscribeKey("busTest2", "id-dupe")
	ch2c := bus.SubscribeKey("busTest2", "id-dupe")

	ch1, _ := bus.SubscribeType("busTest1")
	ch2, _ := bus.SubscribeType("busTest2")

	// Overlapping IDs but not types
	bus.Announce("busTest1", &busTest{
		Metadata: metadata.Metadata{
			ID: "id-overlap",
		},
	})

	msg := <-ch1a
	require.Equal(t, "id-overlap", msg.(*busTest).ID)

	msg = <-ch1
	require.Equal(t, "id-overlap", msg.(*busTest).ID)

	select {
	case msg := <-ch2a:
		t.Errorf("%+v", msg)
	case msg := <-ch2:
		t.Errorf("%+v", msg)
	default:
	}

	bus.Announce("busTest2", &busTest{
		Metadata: metadata.Metadata{
			ID: "id-overlap",
		},
	})

	select {
	case msg := <-ch1a:
		t.Errorf("%+v", msg)
	case msg := <-ch1:
		t.Errorf("%+v", msg)
	default:
	}

	msg = <-ch2a
	require.Equal(t, "id-overlap", msg.(*busTest).ID)

	msg = <-ch2
	require.Equal(t, "id-overlap", msg.(*busTest).ID)

	bus.Announce("busTest2", &busTest{
		Metadata: metadata.Metadata{
			ID: "id-dupe",
		},
	})

	msg = <-ch2b
	require.Equal(t, "id-dupe", msg.(*busTest).ID)

	msg = <-ch2c
	require.Equal(t, "id-dupe", msg.(*busTest).ID)

	msg = <-ch2
	require.Equal(t, "id-dupe", msg.(*busTest).ID)
}

func TestBusDelete(t *testing.T) {
	t.Parallel()

	bus := bus.NewBus()

	keyChan := bus.SubscribeKey("busTest", "id1")

	typeChan, delChan := bus.SubscribeType("busTest")

	bus.Announce("busTest", &busTest{
		Metadata: metadata.Metadata{
			ID: "id1",
		},
	})

	msg := <-keyChan
	require.Equal(t, "id1", msg.(*busTest).ID)

	msg = <-typeChan
	require.Equal(t, "id1", msg.(*busTest).ID)

	bus.Delete("busTest", "id1")

	_, ok := <-keyChan
	require.False(t, ok)

	id := <-delChan
	require.Equal(t, "id1", id)
}

type busTest struct {
	metadata.Metadata
}
