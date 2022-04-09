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
	ch1a := bus.SubscribeKey("busTest1", &busTest{
		Metadata: metadata.Metadata{
			Id: "id-overlap",
		},
	})

	ch2a := bus.SubscribeKey("busTest2", &busTest{
		Metadata: metadata.Metadata{
			Id: "id-overlap",
		},
	})

	ch2b := bus.SubscribeKey("busTest2", &busTest{
		Metadata: metadata.Metadata{
			Id: "id-dupe",
		},
	})

	ch2c := bus.SubscribeKey("busTest2", &busTest{
		Metadata: metadata.Metadata{
			Id: "id-dupe",
		},
	})

	ch1, _ := bus.SubscribeType("busTest1")
	ch2, _ := bus.SubscribeType("busTest2")

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

	msg = <-ch1
	if msg.(*busTest).Id != "id-overlap" {
		t.Errorf("%+v", msg)
	}

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
	if msg.(*busTest).Id != "id-overlap" {
		t.Errorf("%+v", msg)
	}

	msg = <-ch2
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

	msg = <-ch2
	if msg.(*busTest).Id != "id-dupe" {
		t.Errorf("%+v", msg)
	}
}

func TestBusDelete(t *testing.T) {
	t.Parallel()

	bus := NewBus()

	keyChan := bus.SubscribeKey("busTest", &busTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	})

	typeChan, delChan := bus.SubscribeType("busTest")

	bus.Announce("busTest", &busTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	})

	msg := <-keyChan
	if msg.(*busTest).Id != "id1" {
		t.Errorf("%+v", msg)
	}

	msg = <-typeChan
	if msg.(*busTest).Id != "id1" {
		t.Errorf("%+v", msg)
	}

	bus.Delete("busTest", &busTest{
		Metadata: metadata.Metadata{
			Id: "id1",
		},
	})

	msg, ok := <-keyChan
	if ok {
		t.Errorf("%+v", msg)
	}

	msg = <-delChan
	if msg.(*busTest).Id != "id1" {
		t.Errorf("%+v", msg)
	}
}

type busTest struct {
	metadata.Metadata
}
