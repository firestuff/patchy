package metadata

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
)

type Metadata struct {
	ID         string `json:"id"`
	ETag       string `json:"etag"`
	Generation int64  `json:"generation"`
}

func HasMetadata(obj any) bool {
	return getMetadataField(obj).IsValid()
}

func GetMetadata(obj any) *Metadata {
	return getMetadataField(obj).Addr().Interface().(*Metadata)
}

func ClearMetadata(obj any) {
	SetMetadata(obj, &Metadata{})
}

func SetMetadata(obj any, md *Metadata) {
	getMetadataField(obj).Set(reflect.ValueOf(*md))
}

func (m *Metadata) GetSafeID() string {
	return GetSafeID(m.ID)
}

func GetSafeID(id string) string {
	// TODO: Make this an hmac to prevent partial collision DoS attacks
	h := sha256.New()
	h.Write([]byte(id))

	return hex.EncodeToString(h.Sum(nil))
}

func getMetadataField(obj any) reflect.Value {
	v := reflect.ValueOf(obj)

	return reflect.Indirect(v).FieldByName("Metadata")
}
