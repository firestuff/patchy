package metadata

import "encoding/hex"
import "fmt"
import "reflect"

type Metadata struct {
	Id     string `json:"id"`
	Sha256 string `json:"sha256"`
}

func GetMetadata(obj any) *Metadata {
	return getMetadataField(obj).Addr().Interface().(*Metadata)
}

func ClearMetadata(obj any) {
	getMetadataField(obj).Set(reflect.ValueOf(Metadata{}))
}

func (m *Metadata) GetSafeId() string {
	return hex.EncodeToString([]byte(m.Id))
}

func (m *Metadata) GetKey(t string) string {
	return fmt.Sprintf("%s:%s", t, m.GetSafeId())
}

func getMetadataField(obj any) reflect.Value {
	v := reflect.ValueOf(obj)
	return reflect.Indirect(v).FieldByName("Metadata")
}
