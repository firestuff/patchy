package path

import "time"

func Equal(obj any, path string, v1Str string) (bool, error) {
	return op(obj, path, v1Str, equal)
}

func equal(v1, v2 any) bool {
	switch v2t := v2.(type) {
	case time.Time:
		tm := v1.(*timeVal)
		return tm.time.Equal(v2t.Truncate(tm.precision))

	default:
		return v1 == v2
	}
}
