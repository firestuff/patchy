package bus

import "testing"

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
	ch1a := bus.Subscribe("busTest1", &busTest{
		Metadata: metadata.Metadata{
			Id: "id-overlap",
		},
	})

	ch2a := bus.Subscribe("busTest2", &busTest{
		Metadata: metadata.Metadata{
			Id: "id-overlap",
		},
	})

	ch2b := bus.Subscribe("busTest2", &busTest{
		Metadata: metadata.Metadata{
			Id: "id-dupe",
		},
	})

	ch2c := bus.Subscribe("busTest2", &busTest{
		Metadata: metadata.Metadata{
			Id: "id-dupe",
		},
	})

	// Overlapping IDs but not types
	bus.Announce("busTest1", &busTest{
		Metadata: metadata.Metadata{
			Id: "id-overlap",
		},
	})

	msg := <-ch1a
	if msg.(*busTest).Id != "id-overlap" {
		t.Errorf("%+v", msg)
	}

	select {
	case msg := <-ch2a:
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
	default:
	}

	msg = <-ch2a
	if msg.(*busTest).Id != "id-overlap" {
		t.Errorf("%+v", msg)
	}

	bus.Announce("busTest2", &busTest{
		Metadata: metadata.Metadata{
			Id: "id-dupe",
		},
	})

	msg = <-ch2b
	if msg.(*busTest).Id != "id-dupe" {
		t.Errorf("%+v", msg)
	}

	msg = <-ch2c
	if msg.(*busTest).Id != "id-dupe" {
		t.Errorf("%+v", msg)
	}
}

func TestBusDelete(t *testing.T) {
	t.Parallel()

	bus := NewBus()

	ch := bus.Subscribe("busTest", &busTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	})

	bus.Announce("busTest", &busTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	})

	msg := <-ch
	if msg.(*busTest).Id != "id1" {
		t.Errorf("%+v", msg)
	}

	bus.Delete("busTest", &busTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	})

	msg, ok := <-ch
	if ok {
		t.Errorf("%+v", msg)
	}
}

type busTest struct {
	metadata.Metadata
}
