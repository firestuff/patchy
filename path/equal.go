package path

import "time"

func Equal(obj any, path string, matchStr string) (bool, error) {
	return op(obj, path, matchStr, equal)
}

func equal(obj, match any, _ string) bool {
	switch objt := obj.(type) {
	case time.Time:
		tm := match.(*timeVal)

		// TODO: Replace Truncate() with a timezone-aware version
		return tm.time.Equal(objt.Truncate(tm.precision))

	default:
		return obj == match
	}
}
