package storebus

import "os"
import "testing"

func TestStoreBus(t *testing.T) {
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

	out2 := (<-ch).(*storeBusTest)
	if out2.Opaque != "foo" {
		t.Errorf("%+v", out2)
	}

	sb.Write(&storeBusTest{
		Id:     "id1",
		Opaque: "bar",
	})

	out3 := (<-ch).(*storeBusTest)
	if out3.Opaque != "bar" {
		t.Errorf("%+v", out3)
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
