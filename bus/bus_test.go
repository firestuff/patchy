package bus

import "testing"

import "github.com/stretchr/testify/require"

import "github.com/firestuff/patchy/metadata"

func TestBus(t *testing.T) {
	t.Parallel()

	bus := NewBus()

	// Announce with no subscribers
	bus.Announce("busTest1", &busTest{
		Metadata: metadata.Metadata{
			Id: "id-nosub",
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
			Id: "id-overlap",
		},
	})

	msg := <-ch1a
	require.Equal(t, "id-overlap", msg.(*busTest).Id)

	msg = <-ch1
	require.Equal(t, "id-overlap", msg.(*busTest).Id)

	select {
	case msg := <-ch2a:
		t.Errorf("%+v", msg)
	case msg := <-ch2:
		t.Errorf("%+v", msg)
	default:
	}

	bus.Announce("busTest2", &busTest{
		Metadata: metadata.Metadata{
			Id: "id-overlap",
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
	require.Equal(t, "id-overlap", msg.(*busTest).Id)

	msg = <-ch2
	require.Equal(t, "id-overlap", msg.(*busTest).Id)

	bus.Announce("busTest2", &busTest{
		Metadata: metadata.Metadata{
			Id: "id-dupe",
		},
	})

	msg = <-ch2b
	require.Equal(t, "id-dupe", msg.(*busTest).Id)

	msg = <-ch2c
	require.Equal(t, "id-dupe", msg.(*busTest).Id)

	msg = <-ch2
	require.Equal(t, "id-dupe", msg.(*busTest).Id)
}

func TestBusDelete(t *testing.T) {
	t.Parallel()

	bus := NewBus()

	keyChan := bus.SubscribeKey("busTest", "id1")

	typeChan, delChan := bus.SubscribeType("busTest")

	bus.Announce("busTest", &busTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	})

	msg := <-keyChan
	require.Equal(t, "id1", msg.(*busTest).Id)

	msg = <-typeChan
	require.Equal(t, "id1", msg.(*busTest).Id)

	bus.Delete("busTest", "id1")

	msg, ok := <-keyChan
	require.False(t, ok)

	id := <-delChan
	require.Equal(t, "id1", id)
}

type busTest struct {
	metadata.Metadata
}
