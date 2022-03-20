package storebus

import "encoding/hex"
import "fmt"

type Object interface {
	GetType() string
	GetId() string
}

func ObjectSafeId(obj Object) string {
	return hex.EncodeToString([]byte(obj.GetId()))
}

func ObjectKey(obj Object) string {
	return fmt.Sprintf("%s:%s", obj.GetType(), ObjectSafeId(obj))
}
