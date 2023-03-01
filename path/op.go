package path

import (
	"strings"
)

func op(obj any, path string, matchStr string, cb func(any, any, string) bool) (bool, error) {
	objVal, err := getAny(obj, path)
	if err != nil {
		return false, err
	}

	matchVal, err := parse(matchStr, objVal)
	if err != nil {
		return false, err
	}

	if isSlice(objVal) {
		return anyTrue(objVal, func(x any) bool { return cb(x, matchVal, matchStr) }), nil
	}

	return cb(objVal, matchVal, matchStr), nil
}

func opList(obj any, path string, matchStr string, cb func(any, any, string) bool) (bool, error) {
	objVal, err := getAny(obj, path)
	if err != nil {
		return false, err
	}

	if objVal == nil {
		return false, nil
	}

	// TODO: Store per-item matchStr
	matchVal := []any{}

	for _, matchPart := range strings.Split(matchStr, ",") {
		matchTmp, err := parse(matchPart, objVal)
		if err != nil {
			return false, err
		}

		matchVal = append(matchVal, matchTmp)
	}

	return anyTrue(matchVal, func(y any) bool {
		if isSlice(objVal) {
			return anyTrue(objVal, func(x any) bool { return cb(x, y, matchStr) })
		}

		return cb(objVal, y, matchStr)
	}), nil
}
