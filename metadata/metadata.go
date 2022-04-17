package metadata

import "crypto/sha256"
import "encoding/hex"
import "reflect"

type Metadata struct {
	Id   string `json:"id"`
	ETag string `json:"etag"`
}

func GetMetadata(obj any) *Metadata {
	return getMetadataField(obj).Addr().Interface().(*Metadata)
}

func ClearMetadata(obj any) {
	getMetadataField(obj).Set(reflect.ValueOf(Metadata{}))
}

func (m *Metadata) GetSafeId() string {
	return GetSafeId(m.Id)
}

func GetSafeId(id string) string {
	// TODO: Make this an hmac to prevent partial collision DoS attacks
	h := sha256.New()
	h.Write([]byte(id))
	return hex.EncodeToString(h.Sum(nil))
}

func getMetadataField(obj any) reflect.Value {
	v := reflect.ValueOf(obj)
	return reflect.Indirect(v).FieldByName("Metadata")
}
