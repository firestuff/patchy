package path

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"github.com/gopatchy/jsrest"
	"golang.org/x/exp/slices"
)

type WalkCallback func(string, []string, reflect.StructField)

var (
	TimeTimeType  = reflect.TypeOf(time.Time{})
	CivilDateType = reflect.TypeOf(civil.Date{})

	ErrNotAStruct       = errors.New("not a struct")
	ErrUnknownFieldName = errors.New("unknown field name")
)

func Get(obj any, path string) (any, error) {
	v, err := GetValue(reflect.ValueOf(obj), path)
	if err != nil {
		return nil, err
	}

	return v.Interface(), nil
}

func GetValue(v reflect.Value, path string) (reflect.Value, error) {
	parts := strings.Split(path, ".")
	return getRecursive(v, parts, []string{})
}

func getRecursive(v reflect.Value, parts []string, prev []string) (reflect.Value, error) {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v = reflect.Zero(v.Type().Elem())
		} else {
			v = reflect.Indirect(v)
		}
	}

	if len(parts) == 0 {
		return v, nil
	}

	if v.Kind() != reflect.Struct {
		return reflect.Value{}, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", strings.Join(prev, "."), ErrNotAStruct)
	}

	part := parts[0]

	sub, found := getField(v, part)
	if !found {
		return reflect.Value{}, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", errorPath(prev, part), ErrUnknownFieldName)
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
	return SetValue(reflect.ValueOf(obj), path, val)
}

func SetValue(v reflect.Value, path string, val string) error {
	parts := strings.Split(path, ".")
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

func List(obj any) []string {
	return ListType(reflect.TypeOf(obj))
}

func ListType(t reflect.Type) []string {
	list := []string{}

	WalkType(t, func(path string, _ []string, field reflect.StructField) {
		t := MaybeIndirectType(field.Type)
		if t.Kind() == reflect.Struct && t != TimeTimeType && t != CivilDateType {
			return
		}

		list = append(list, path)
	})

	sort.Strings(list)

	return list
}

func GetFieldType(t reflect.Type, path string) reflect.Type {
	parts := strings.Split(path, ".")

	for _, part := range parts {
		field, found := getStructField(MaybeIndirectType(t), part)
		if !found {
			return nil
		}

		t = field.Type
	}

	return t
}

func FindTagValueType(t reflect.Type, key, value string) (string, bool) {
	ret := ""

	WalkType(t, func(path string, _ []string, field reflect.StructField) {
		tag, found := field.Tag.Lookup(key)
		if !found {
			return
		}

		parts := strings.Split(tag, ",")

		if slices.Contains(parts, value) {
			ret = path
		}
	})

	return ret, ret != ""
}

func Walk(obj any, cb WalkCallback) {
	WalkType(reflect.TypeOf(obj), cb)
}

func WalkType(t reflect.Type, cb WalkCallback) {
	walkRecursive(MaybeIndirectType(t), cb, []string{})
}

func walkRecursive(t reflect.Type, cb WalkCallback, prev []string) {
	for i := 0; i < t.NumField(); i++ {
		sub := t.Field(i)

		newPrev := []string{}
		newPrev = append(newPrev, prev...)

		if !sub.Anonymous {
			newPrev = append(newPrev, FieldName(sub))
		}

		t := MaybeIndirectType(sub.Type)

		if len(newPrev) > 0 {
			cb(strings.Join(newPrev, "."), newPrev, sub)
		}

		if t.Kind() == reflect.Struct && t != TimeTimeType && t != CivilDateType {
			walkRecursive(t, cb, newPrev)
		}
	}
}

func getStructField(t reflect.Type, name string) (reflect.StructField, bool) {
	name = strings.ToLower(name)

	return t.FieldByNameFunc(func(iterName string) bool {
		iterField, iterOK := t.FieldByName(iterName)
		if !iterOK {
			panic(iterName)
		}

		return strings.ToLower(FieldName(iterField)) == name
	})
}

func errorPath(prev []string, part string) string {
	if len(prev) == 0 {
		return part
	}

	return fmt.Sprintf("%s.%s", strings.Join(prev, "."), part)
}

func FieldName(field reflect.StructField) string {
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

func MaybeIndirectType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Pointer {
		return t.Elem()
	}

	return t
}
