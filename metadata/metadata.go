package metadata

import "encoding/hex"
import "fmt"
import "reflect"

type Metadata struct {
	Id string `json:"id"`
}

func GetMetadata(obj interface{}) *Metadata {
	return getMetadataField(obj).Addr().Interface().(*Metadata)
}

func ClearMetadata(obj interface{}) {
	getMetadataField(obj).Set(reflect.ValueOf(Metadata{}))
}

func (m *Metadata) GetSafeId() string {
	return hex.EncodeToString([]byte(m.Id))
}

func (m *Metadata) GetKey(t string) string {
	return fmt.Sprintf("%s:%s", t, m.GetSafeId())
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
