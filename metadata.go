package storebus

import "encoding/hex"
import "fmt"
import "reflect"

type Metadata struct {
	Id string `json:"id"`
}

func getMetadata(obj interface{}) *Metadata {
	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}

	m := v.FieldByName("Metadata")
	if !m.IsValid() {
		panic(fmt.Sprintf("Metadata field missing in %s", v.Type()))
	}

	return m.Addr().Interface().(*Metadata)
}

func (m *Metadata) getSafeId() string {
	return hex.EncodeToString([]byte(m.Id))
}

func (m *Metadata) getKey(t string) string {
	return fmt.Sprintf("%s:%s", t, m.getSafeId())
}
