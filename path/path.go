package path

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"github.com/firestuff/patchy/jsrest"
)

var (
	ErrNotAStruct       = errors.New("not a struct")
	ErrUnknownFieldName = errors.New("unknown field name")
)

func Get(obj any, path string) (any, error) {
	parts := strings.Split(path, ".")
	v := reflect.ValueOf(obj)

	return getRecursive(v, parts, []string{})
}

func getRecursive(v reflect.Value, parts []string, prev []string) (any, error) {
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

	sub, found := getField(v, part)
	if !found {
		return nil, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", errorPath(prev, part), ErrUnknownFieldName)
	}

	newPrev := []string{}
	newPrev = append(newPrev, prev...)
	newPrev = append(newPrev, part)

	return getRecursive(sub, parts[1:], newPrev)
}

func getField(v reflect.Value, name string) (reflect.Value, bool) {
	field, found := getStructField(v.Type(), name)
	if !found {
		return reflect.Value{}, false
	}

	return v.FieldByName(field.Name), true
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

	sub, found := getField(v, part)
	if !found {
		return jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", errorPath(prev, part), ErrUnknownFieldName)
	}

	newPrev := []string{}
	newPrev = append(newPrev, prev...)
	newPrev = append(newPrev, part)

	return setRecursive(sub, parts[1:], newPrev, val)
}

func List(obj any) ([]string, error) {
	return ListType(reflect.TypeOf(obj))
}

func ListType(t reflect.Type) ([]string, error) {
	list, err := listRecursive(t, []string{}, []string{})
	if err != nil {
		return nil, err
	}

	sort.Strings(list)

	return list, nil
}

func listRecursive(t reflect.Type, prev []string, list []string) ([]string, error) {
	var err error

	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct || t == reflect.TypeOf(time.Time{}) || t == reflect.TypeOf(civil.Date{}) {
		if len(prev) > 0 {
			list = append(list, strings.Join(prev, "."))
		}

		return list, nil
	}

	for i := 0; i < t.NumField(); i++ {
		sub := t.Field(i)

		newPrev := []string{}
		newPrev = append(newPrev, prev...)

		if !sub.Anonymous {
			newPrev = append(newPrev, fieldName(sub)) //nolint:makezero
		}

		list, err = listRecursive(sub.Type, newPrev, list)
		if err != nil {
			return nil, err
		}
	}

	return list, nil
}

func GetFieldType(t reflect.Type, path string) reflect.Type {
	parts := strings.Split(path, ".")

	for _, part := range parts {
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}

		field, found := getStructField(t, part)
		if !found {
			return nil
		}

		t = field.Type
	}

	return t
}

func getStructField(t reflect.Type, name string) (reflect.StructField, bool) {
	name = strings.ToLower(name)

	return t.FieldByNameFunc(func(iterName string) bool {
		iterField, iterOK := t.FieldByName(iterName)
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
}

func errorPath(prev []string, part string) string {
	if len(prev) == 0 {
		return part
	}

	return fmt.Sprintf("%s.%s", strings.Join(prev, "."), part)
}

func fieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag != "" {
		if tag == "-" {
			return ""
		}

		parts := strings.SplitN(tag, ",", 2)

		return parts[0]
	}

	return field.Name
}
