package bus_test

import (
	"context"
	"testing"

	"github.com/firestuff/patchy/bus"
	"github.com/firestuff/patchy/metadata"
	"github.com/firestuff/patchy/view"
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

	ev1a, err := view.NewEphemeralView[any](context.TODO(), nil)
	require.Nil(t, err)
	<-ev1a.Chan()

	ev2a, err := view.NewEphemeralView[any](context.TODO(), nil)
	require.Nil(t, err)
	<-ev2a.Chan()

	ev2b, err := view.NewEphemeralView[any](context.TODO(), nil)
	require.Nil(t, err)
	<-ev2b.Chan()

	ev2c, err := view.NewEphemeralView[any](context.TODO(), nil)
	require.Nil(t, err)
	<-ev2c.Chan()

	ev1, err := view.NewEphemeralView[any](context.TODO(), nil)
	require.Nil(t, err)
	<-ev1.Chan()

	ev2, err := view.NewEphemeralView[any](context.TODO(), nil)
	require.Nil(t, err)
	<-ev2.Chan()

	// Complex subscription layout
	bus.SubscribeKey("busTest1", "id-overlap", ev1a)
	bus.SubscribeKey("busTest2", "id-overlap", ev2a)
	bus.SubscribeKey("busTest2", "id-dupe", ev2b)
	bus.SubscribeKey("busTest2", "id-dupe", ev2c)

	ch1, _ := bus.SubscribeType("busTest1")
	ch2, _ := bus.SubscribeType("busTest2")

	// Overlapping IDs but not types
	bus.Announce("busTest1", &busTest{
		Metadata: metadata.Metadata{
			ID: "id-overlap",
		},
	})

	msg := <-ev1a.Chan()
	require.Equal(t, "id-overlap", msg.(*busTest).ID)

	msg = <-ch1
	require.Equal(t, "id-overlap", msg.(*busTest).ID)

	select {
	case msg := <-ev2a.Chan():
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
	case msg := <-ev1a.Chan():
		t.Errorf("%+v", msg)
	case msg := <-ch1:
		t.Errorf("%+v", msg)
	default:
	}

	msg = <-ev2a.Chan()
	require.Equal(t, "id-overlap", msg.(*busTest).ID)

	msg = <-ch2
	require.Equal(t, "id-overlap", msg.(*busTest).ID)

	bus.Announce("busTest2", &busTest{
		Metadata: metadata.Metadata{
			ID: "id-dupe",
		},
	})

	msg = <-ev2b.Chan()
	require.Equal(t, "id-dupe", msg.(*busTest).ID)

	msg = <-ev2c.Chan()
	require.Equal(t, "id-dupe", msg.(*busTest).ID)

	msg = <-ch2
	require.Equal(t, "id-dupe", msg.(*busTest).ID)
}

func TestBusDelete(t *testing.T) {
	t.Parallel()

	bus := bus.NewBus()

	ev, err := view.NewEphemeralView[any](context.TODO(), nil)
	require.Nil(t, err)
	<-ev.Chan()
	bus.SubscribeKey("busTest", "id1", ev)

	typeChan, delChan := bus.SubscribeType("busTest")

	bus.Announce("busTest", &busTest{
		Metadata: metadata.Metadata{
			ID: "id1",
		},
	})

	msg := <-ev.Chan()
	require.Equal(t, "id1", msg.(*busTest).ID)

	msg = <-typeChan
	require.Equal(t, "id1", msg.(*busTest).ID)

	bus.Delete("busTest", "id1")

	_, ok := <-ev.Chan()
	require.False(t, ok)

	id := <-delChan
	require.Equal(t, "id1", id)
}

type busTest struct {
	metadata.Metadata
}
