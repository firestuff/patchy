package path

import "time"

import "cloud.google.com/go/civil"

func LessEqual(obj any, path string, v1Str string) (bool, error) {
	return op(obj, path, v1Str, lessEqual)
}

func lessEqual(v1, v2 any) bool {
	switch v2t := v2.(type) {
	case int:
		return v2t <= v1.(int)

	case int64:
		return v2t <= v1.(int64)

	case uint:
		return v2t <= v1.(uint)

	case uint64:
		return v2t <= v1.(uint64)

	case float32:
		return v2t <= v1.(float32)

	case float64:
		return v2t <= v1.(float64)

	case string:
		return v2t <= v1.(string)

	case bool:
		return v2t == false || v2t == v1.(bool)

	case time.Time:
		tm := v1.(*timeVal)
		trunc := v2t.Truncate(tm.precision)
		return trunc.Equal(tm.time) || trunc.Before(tm.time)

	case civil.Date:
		return v2t == v1.(civil.Date) || v2t.Before(v1.(civil.Date))

	default:
		panic(v2)
	}
}
