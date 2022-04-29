package path

import "time"

func Equal(obj any, path string, v1Str string) (bool, error) {
	v2, err := getAny(obj, path)
	if err != nil {
		return false, err
	}
	if v2 == nil {
		return false, nil
	}

	v1, err := parse(v1Str, v2)
	if err != nil {
		return false, err
	}

	return equal(v1, v2), nil
}

func equal(v1 any, v2 any) bool {
	if v2 == nil {
		return false
	}

	if isSlice(v2) {
		return anyTrue(v2, func(x any) bool { return equal(v1, x) })
	}

	switch v2t := v2.(type) {
	case time.Time:
		tm := v1.(*timeVal)
		return tm.time.Equal(v2t.Truncate(tm.precision))

	default:
		return v1 == v2
	}
}
