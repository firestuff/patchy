package api

import "reflect"

func Merge(to, from any) {
	MergeValue(reflect.ValueOf(to), reflect.ValueOf(from))
}

func MergeValue(to, from reflect.Value) {
	to = maybeIndirect(to)
	from = maybeIndirect(from)

	for i := 0; i < to.NumField(); i++ {
		toField := to.Field(i)
		fromField := from.Field(i)

		if fromField.IsZero() {
			continue
		}

		if maybeIndirect(fromField).Kind() == reflect.Struct {
			MergeValue(toField, fromField)
			continue
		}

		toField.Set(fromField)
	}
}

func maybeIndirect(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}

	return v
}
