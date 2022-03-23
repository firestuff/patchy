package storebus

import "os"
import "testing"

func TestStoreBus(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	sb := NewStoreBus(dir)

	err = sb.Write(&storeBusTest{
		Id:     "id1",
		Opaque: "foo",
	})
	if err != nil {
		t.Fatal(err)
	}

	out1 := &storeBusTest{
		Id: "id1",
	}

	err = sb.Read(out1)
	if err != nil {
		t.Fatal(err)
	}

	if out1.Opaque != "foo" {
		t.Errorf("%+v", out1)
	}

	ch := sb.Subscribe(&storeBusTest{
		Id: "id1",
	})

	sb.Write(&storeBusTest{
		Id:     "id1",
		Opaque: "bar",
	})

	out3 := (<-ch).(*storeBusTest)
	if out3.Opaque != "bar" {
		t.Errorf("%+v", out3)
	}
}

func TestStoreBusDelete(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	sb := NewStoreBus(dir)

	ch := sb.Subscribe(&storeBusTest{
		Id: "id1",
	})

	sb.Write(&storeBusTest{
		Id:     "id1",
		Opaque: "foo",
	})

	out := (<-ch).(*storeBusTest)
	if out.Opaque != "foo" {
		t.Errorf("%+v", out)
	}

	sb.Delete(&storeBusTest{
		Id: "id1",
	})

	out2, ok := <-ch
	if ok {
		t.Errorf("%+v", out2)
	}
}

type storeBusTest struct {
	Id     string
	Opaque string
}

func (sbt *storeBusTest) GetType() string {
	return "storeBusTest"
}

func (sbt *storeBusTest) GetId() string {
	return sbt.Id
}

func (sbt *storeBusTest) SetId(id string) {
	sbt.Id = id
}
