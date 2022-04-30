package path

import (
	"fmt"
	"reflect"
	"strings"
)

var (
	errNotAStruct       = fmt.Errorf("not a struct")
	errUnknownFieldName = fmt.Errorf("unknown field name")
)

func getAny(obj any, path string) (any, error) {
	parts := strings.Split(path, ".")
	v := reflect.ValueOf(obj)

	return getAnyRecursive(v, parts, []string{})
}

func getAnyRecursive(v reflect.Value, parts []string, prev []string) (any, error) {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			// Not an error, just a lack of a result
			return nil, nil
		}

		v = reflect.Indirect(v)
	}

	if len(parts) == 0 {
		return v.Interface(), nil
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s: %w", strings.Join(prev, "."), errNotAStruct)
	}

	part := parts[0]

	sub := getField(v, part)
	if !sub.IsValid() {
		return nil, fmt.Errorf("%s: %w", errorPath(prev, part), errUnknownFieldName)
	}

	newPrev := []string{}
	copy(newPrev, prev)
	newPrev = append(newPrev, part)

	return getAnyRecursive(sub, parts[1:], newPrev)
}

func getField(v reflect.Value, name string) reflect.Value {
	name = strings.ToLower(name)
	typ := v.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Name

		tag := field.Tag.Get("json")
		if tag != "" {
			if tag == "-" {
				continue
			}

			parts := strings.SplitN(tag, ",", 2)
			fieldName = parts[0]
		}

		if strings.ToLower(fieldName) == name {
			return v.Field(i)
		}
	}

	return reflect.Value{}
}

func errorPath(prev []string, part string) string {
	if len(prev) == 0 {
		return part
	}

	return fmt.Sprintf("%s.%s", strings.Join(prev, "."), part)
}
