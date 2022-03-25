package metadata

import "encoding/hex"
import "fmt"
import "reflect"

type Metadata struct {
	Id string `json:"id"`
}

func GetMetadata(obj any) *Metadata {
	return GetMetadataField(obj).Addr().Interface().(*Metadata)
}

func ClearMetadata(obj any) {
	GetMetadataField(obj).Set(reflect.ValueOf(Metadata{}))
}

func (m *Metadata) GetSafeId() string {
	return hex.EncodeToString([]byte(m.Id))
}

func (m *Metadata) GetKey(t string) string {
	return fmt.Sprintf("%s:%s", t, m.GetSafeId())
}

func GetMetadataField(obj any) reflect.Value {
	return maybeIndirect(obj).FieldByName("Metadata")
}

func maybeIndirect(obj any) reflect.Value {
	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}

	return v
}
