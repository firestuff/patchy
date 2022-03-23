package storebus

import "encoding/hex"
import "fmt"
import "reflect"

type Metadata struct {
	Id string `json:"id"`
}

func getMetadata(obj interface{}) *Metadata {
	return getMetadataField(obj).Addr().Interface().(*Metadata)
}

func clearMetadata(obj interface{}) {
	getMetadataField(obj).Set(reflect.ValueOf(Metadata{}))
}

func getMetadataField(obj interface{}) reflect.Value {
	v := maybeIndirect(obj)

	m := v.FieldByName("Metadata")
	if !m.IsValid() {
		panic(fmt.Sprintf("Metadata field missing in %s", v.Type()))
	}

	return m
}

func maybeIndirect(obj interface{}) reflect.Value {
	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}

	return v
}

func (m *Metadata) getSafeId() string {
	return hex.EncodeToString([]byte(m.Id))
}

func (m *Metadata) getKey(t string) string {
	return fmt.Sprintf("%s:%s", t, m.getSafeId())
}
