package storebus

import "testing"

func TestBus(t *testing.T) {
	t.Parallel()

	bus := NewBus()

	// Announce with no subscribers
	bus.Announce(&busTest1{
		Id: "id-nosub",
	})

	// Complex subscription layout
	ch1a := bus.Subscribe(&busTest1{
		Id: "id-overlap",
	})

	ch2a := bus.Subscribe(&busTest2{
		Id: "id-overlap",
	})

	ch2b := bus.Subscribe(&busTest2{
		Id: "id-dupe",
	})

	ch2c := bus.Subscribe(&busTest2{
		Id: "id-dupe",
	})

	// Overlapping IDs but not types
	bus.Announce(&busTest1{
		Id: "id-overlap",
	})

	msg := <-ch1a
	if msg.(*busTest1).Id != "id-overlap" {
		t.Errorf("%+v", msg)
	}

	select {
	case msg := <-ch2a:
		t.Errorf("%+v", msg)
	default:
	}

	bus.Announce(&busTest2{
		Id: "id-overlap",
	})

	select {
	case msg := <-ch1a:
		t.Errorf("%+v", msg)
	default:
	}

	msg = <-ch2a
	if msg.(*busTest2).Id != "id-overlap" {
		t.Errorf("%+v", msg)
	}

	bus.Announce(&busTest2{
		Id: "id-dupe",
	})

	msg = <-ch2b
	if msg.(*busTest2).Id != "id-dupe" {
		t.Errorf("%+v", msg)
	}

	msg = <-ch2c
	if msg.(*busTest2).Id != "id-dupe" {
		t.Errorf("%+v", msg)
	}
}

type busTest1 struct {
	Id string
}

func (bt *busTest1) GetType() string {
	return "busTest1"
}

func (bt *busTest1) GetId() string {
	return bt.Id
}

func (bt *busTest1) SetId(id string) {
	bt.Id = id
}

type busTest2 struct {
	Id string
}

func (bt *busTest2) GetType() string {
	return "busTest2"
}

func (bt *busTest2) GetId() string {
	return bt.Id
}

func (bt *busTest2) SetId(id string) {
	bt.Id = id
}
