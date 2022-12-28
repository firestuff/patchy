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

	ev1a := view.NewEphemeralViewEmpty[any](context.Background())
	ev2a := view.NewEphemeralViewEmpty[any](context.Background())
	ev2b := view.NewEphemeralViewEmpty[any](context.Background())
	ev2c := view.NewEphemeralViewEmpty[any](context.Background())
	evt1 := view.NewEphemeralViewEmpty[any](context.Background())
	evt2 := view.NewEphemeralViewEmpty[any](context.Background())

	// Complex subscription layout
	bus.SubscribeKey("busTest1", "id-overlap", ev1a)
	bus.SubscribeKey("busTest2", "id-overlap", ev2a)
	bus.SubscribeKey("busTest2", "id-dupe", ev2b)
	bus.SubscribeKey("busTest2", "id-dupe", ev2c)

	bus.SubscribeType("busTest1", evt1)
	bus.SubscribeType("busTest2", evt2)

	// Overlapping IDs but not types
	bus.Announce("busTest1", &busTest{
		Metadata: metadata.Metadata{
			ID: "id-overlap",
		},
	})

	msg := <-ev1a.Chan()
	require.Equal(t, "id-overlap", msg.(*busTest).ID)

	msg = <-evt1.Chan()
	require.Equal(t, "id-overlap", msg.(*busTest).ID)

	select {
	case msg := <-ev2a.Chan():
		t.Errorf("%+v", msg)
	case msg := <-evt2.Chan():
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
	case msg := <-evt1.Chan():
		t.Errorf("%+v", msg)
	default:
	}

	msg = <-ev2a.Chan()
	require.Equal(t, "id-overlap", msg.(*busTest).ID)

	msg = <-evt2.Chan()
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

	msg = <-evt2.Chan()
	require.Equal(t, "id-dupe", msg.(*busTest).ID)
}

func TestBusDelete(t *testing.T) {
	t.Parallel()

	bus := bus.NewBus()

	ev, err := view.NewEphemeralView[any](context.Background(), nil)
	require.Nil(t, err)
	<-ev.Chan()
	bus.SubscribeKey("busTest", "id1", ev)

	evt, err := view.NewEphemeralView[any](context.Background(), nil)
	require.Nil(t, err)
	<-evt.Chan()
	bus.SubscribeType("busTest", evt)

	bus.Announce("busTest", &busTest{
		Metadata: metadata.Metadata{
			ID: "id1",
		},
	})

	msg := <-ev.Chan()
	require.Equal(t, "id1", msg.(*busTest).ID)

	msg = <-evt.Chan()
	require.Equal(t, "id1", msg.(*busTest).ID)

	bus.Delete("busTest", "id1")

	_, ok := <-ev.Chan()
	require.False(t, ok)

	id := (<-evt.Chan()).(string)
	require.Equal(t, "id1", id)
}

type busTest struct {
	metadata.Metadata
}
