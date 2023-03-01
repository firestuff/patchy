package path

import (
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/civil"
)

func HasPrefix(obj any, path string, v1Str string) (bool, error) {
	return op(obj, path, v1Str, hasPrefix)
}

// TODO: These should operate on strings for the match value, not parsed values.
func hasPrefix(v1, v2 any) bool {
	var s1, s2 string

	switch v2t := v2.(type) {
	case int:
		s1 = strconv.FormatInt(int64(v1.(int)), 10)
		s2 = strconv.FormatInt(int64(v2t), 10)

	case int64:
		s1 = strconv.FormatInt(v1.(int64), 10)
		s2 = strconv.FormatInt(v2t, 10)

	case uint:
		s1 = strconv.FormatUint(uint64(v1.(uint)), 10)
		s2 = strconv.FormatUint(uint64(v2t), 10)

	case uint64:
		s1 = strconv.FormatUint(v1.(uint64), 10)
		s2 = strconv.FormatUint(v2t, 10)

	case float32:
		s1 = strconv.FormatFloat(float64(v1.(float32)), 'f', -1, 32)
		s2 = strconv.FormatFloat(float64(v2t), 'f', -1, 32)

	case float64:
		s1 = strconv.FormatFloat(v1.(float64), 'f', -1, 64)
		s2 = strconv.FormatFloat(v2t, 'f', -1, 64)

	case string:
		s1 = v1.(string)
		s2 = v2t

	// These last 3 don't make a whole lot of sense, but implement the best we can
	case bool:
		s1 = strconv.FormatBool(v1.(bool))
		s2 = strconv.FormatBool(v2t)

	case time.Time:
		s1 = v1.(*timeVal).time.String()
		s2 = v2t.String()

	case civil.Date:
		s1 = v1.(civil.Date).String()
		s2 = v2t.String()

	default:
		panic(v2)
	}

	return strings.HasPrefix(s2, s1)
}
