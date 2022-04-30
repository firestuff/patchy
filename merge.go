package patchy

import "reflect"

func Merge(to any, from any) error {
	t := maybeIndirect(to)
	f := maybeIndirect(from)

	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		ff := f.Field(i)

		if ff.IsZero() {
			continue
		}

		tf.Set(ff)
	}

	return nil
}

func maybeIndirect(obj any) reflect.Value {
	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}

	return v
}
