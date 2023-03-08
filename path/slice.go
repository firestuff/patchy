package path

import "reflect"

func isSlice(v any) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}

func anyTrue(v any, cb func(any, int) bool) bool {
	val := reflect.ValueOf(v)

	for i := 0; i < val.Len(); i++ {
		sub := val.Index(i)

		if sub.Kind() == reflect.Pointer && sub.IsNil() {
			continue
		}

		sub = reflect.Indirect(sub)

		if cb(sub.Interface(), i) {
			return true
		}
	}

	return false
}
