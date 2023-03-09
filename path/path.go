package path

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/firestuff/patchy/jsrest"
)

var (
	ErrNotAStruct       = errors.New("not a struct")
	ErrUnknownFieldName = errors.New("unknown field name")
)

func getAny(obj any, path string) (any, error) {
	parts := strings.Split(path, ".")
	v := reflect.ValueOf(obj)

	return getAnyRecursive(v, parts, []string{})
}

func getAnyRecursive(v reflect.Value, parts []string, prev []string) (any, error) {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v = reflect.Zero(v.Type().Elem())
		} else {
			v = reflect.Indirect(v)
		}
	}

	if len(parts) == 0 {
		return v.Interface(), nil
	}

	if v.Kind() != reflect.Struct {
		return nil, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", strings.Join(prev, "."), ErrNotAStruct)
	}

	part := parts[0]

	sub := getField(v, part)
	if !sub.IsValid() {
		return nil, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", errorPath(prev, part), ErrUnknownFieldName)
	}

	newPrev := []string{}
	copy(newPrev, prev)
	newPrev = append(newPrev, part)

	return getAnyRecursive(sub, parts[1:], newPrev)
}

func getField(v reflect.Value, name string) reflect.Value {
	name = strings.ToLower(name)
	typ := v.Type()

	field, ok := typ.FieldByNameFunc(func(iterName string) bool {
		iterField, iterOK := typ.FieldByName(iterName)
		if !iterOK {
			panic(iterName)
		}

		tag := iterField.Tag.Get("json")
		if tag != "" {
			if tag == "-" {
				return false
			}

			parts := strings.SplitN(tag, ",", 2)
			iterName = parts[0]
		}

		return strings.ToLower(iterName) == name
	})

	if ok {
		return v.FieldByName(field.Name)
	}

	return reflect.Value{}
}

func Set(obj any, path string, val string) error {
	parts := strings.Split(path, ".")
	v := reflect.ValueOf(obj)

	return setRecursive(v, parts, []string{}, val)
}

func setRecursive(v reflect.Value, parts []string, prev []string, val string) error {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		v = reflect.Indirect(v)
	}

	if len(parts) == 0 {
		n, err := parse(val, v.Interface())
		if err != nil {
			return err
		}

		if _, ok := n.(*timeVal); ok {
			n = n.(*timeVal).time
		}

		v.Set(reflect.ValueOf(n))

		return nil
	}

	if v.Kind() != reflect.Struct {
		return jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", strings.Join(prev, "."), ErrNotAStruct)
	}

	part := parts[0]

	sub := getField(v, part)
	if !sub.IsValid() {
		return jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", errorPath(prev, part), ErrUnknownFieldName)
	}

	newPrev := []string{}
	copy(newPrev, prev)
	newPrev = append(newPrev, part)

	return setRecursive(sub, parts[1:], newPrev, val)
}

func errorPath(prev []string, part string) string {
	if len(prev) == 0 {
		return part
	}

	return fmt.Sprintf("%s.%s", strings.Join(prev, "."), part)
}
