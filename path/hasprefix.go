package path

import (
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/civil"
)

func HasPrefix(obj any, path string, matchStr string) (bool, error) {
	return op(obj, path, matchStr, hasPrefix)
}

func hasPrefix(obj, match any, matchStr string) bool {
	var objStr string

	switch objt := obj.(type) {
	case int:
		objStr = strconv.FormatInt(int64(objt), 10)

	case int64:
		objStr = strconv.FormatInt(objt, 10)

	case uint:
		objStr = strconv.FormatUint(uint64(objt), 10)

	case uint64:
		objStr = strconv.FormatUint(objt, 10)

	case float32:
		objStr = strconv.FormatFloat(float64(objt), 'f', -1, 32)

	case float64:
		objStr = strconv.FormatFloat(objt, 'f', -1, 64)

	case string:
		objStr = objt

	case bool:
		objStr = strconv.FormatBool(objt)

	case time.Time:
		objStr = objt.String()

	case civil.Date:
		objStr = objt.String()

	default:
		panic(obj)
	}

	return strings.HasPrefix(objStr, matchStr)
}
