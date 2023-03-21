package path

import (
	"strings"
)

func op(obj any, path string, matchStr string, cb func(any, any, string) bool) (bool, error) {
	objVal, err := Get(obj, path)
	if err != nil {
		return false, err
	}

	matchVal, err := parse(matchStr, objVal)
	if err != nil {
		return false, err
	}

	if isSlice(objVal) {
		return anyTrue(objVal, func(x any, _ int) bool { return cb(x, matchVal, matchStr) }), nil
	}

	return cb(objVal, matchVal, matchStr), nil
}

func opList(obj any, path string, matchStr string, cb func(any, any, string) bool) (bool, error) {
	objVal, err := Get(obj, path)
	if err != nil {
		return false, err
	}

	if objVal == nil {
		return false, nil
	}

	matchVal := []any{}
	matchParts := strings.Split(matchStr, ",")

	for _, matchPart := range matchParts {
		matchTmp, err := parse(matchPart, objVal)
		if err != nil {
			return false, err
		}

		matchVal = append(matchVal, matchTmp)
	}

	return anyTrue(matchVal, func(y any, i int) bool {
		str := matchParts[i]

		if isSlice(objVal) {
			return anyTrue(objVal, func(x any, _ int) bool { return cb(x, y, str) })
		}

		return cb(objVal, y, str)
	}), nil
}
