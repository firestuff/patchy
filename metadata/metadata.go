package metadata

import (
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

func getMetadataField(obj any) reflect.Value {
	v := reflect.ValueOf(obj)

	return reflect.Indirect(v).FieldByName("Metadata")
}
