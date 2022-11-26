package path

import (
	"strings"
)

func op(obj any, path string, v1Str string, cb func(any, any) bool) (bool, error) {
	v2, err := getAny(obj, path)
	if err != nil {
		return false, err
	}

	v1, err := parse(v1Str, v2)
	if err != nil {
		return false, err
	}

	if isSlice(v2) {
		return anyTrue(v2, func(x any) bool { return cb(v1, x) }), nil
	}

	return cb(v1, v2), nil
}

func opList(obj any, path string, v1Str string, cb func(any, any) bool) (bool, error) {
	v2, err := getAny(obj, path)
	if err != nil {
		return false, err
	}

	if v2 == nil {
		return false, nil
	}

	v1 := []any{}

	for _, v1Part := range strings.Split(v1Str, ",") {
		v1Tmp, err := parse(v1Part, v2)
		if err != nil {
			return false, err
		}

		v1 = append(v1, v1Tmp)
	}

	return anyTrue(v1, func(y any) bool {
		if isSlice(v2) {
			return anyTrue(v2, func(x any) bool { return cb(y, x) })
		}

		return cb(y, v2)
	}), nil
}
