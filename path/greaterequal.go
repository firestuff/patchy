package path

import (
	"time"

	"cloud.google.com/go/civil"
)

func GreaterEqual(obj any, path string, matchStr string) (bool, error) {
	return op(obj, path, matchStr, greaterEqual)
}

func greaterEqual(obj, match any, _ string) bool {
	switch objt := obj.(type) {
	case int:
		return objt >= match.(int)

	case int64:
		return objt >= match.(int64)

	case uint:
		return objt >= match.(uint)

	case uint64:
		return objt >= match.(uint64)

	case float32:
		return objt >= match.(float32)

	case float64:
		return objt >= match.(float64)

	case string:
		return objt >= match.(string)

	case bool:
		return objt || objt == match.(bool)

	case time.Time:
		tm := match.(*timeVal)
		trunc := objt.Truncate(tm.precision)

		return trunc.Equal(tm.time) || trunc.After(tm.time)

	case civil.Date:
		return objt == match.(civil.Date) || objt.After(match.(civil.Date))

	default:
		panic(obj)
	}
}
