package storebus

import "encoding/hex"
import "fmt"

type Object interface {
	GetId() string
	SetId(string)
}

func ObjectSafeId(obj Object) string {
	return hex.EncodeToString([]byte(obj.GetId()))
}

func ObjectKey(t string, obj Object) string {
	return fmt.Sprintf("%s:%s", t, ObjectSafeId(obj))
}
