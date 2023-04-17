package path

import "reflect"

func Merge(to, from any) {
	MergeValue(reflect.ValueOf(to), reflect.ValueOf(from))
}

func MergeValue(to, from reflect.Value) {
	to = reflect.Indirect(to)
	from = reflect.Indirect(from)

	for i := 0; i < to.NumField(); i++ {
		toField := to.Field(i)
		fromField := from.Field(i)

		if fromField.IsZero() {
			continue
		}

		if reflect.Indirect(fromField).Kind() == reflect.Struct {
			MergeValue(toField, fromField)
			continue
		}

		toField.Set(fromField)
	}
}
