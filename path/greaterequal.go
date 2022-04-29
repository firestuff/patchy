package path

import "fmt"
import "time"

import "cloud.google.com/go/civil"
import "golang.org/x/exp/constraints"

func GreaterEqual(obj any, path string, val1 string) (bool, error) {
	val2, err := getAny(obj, path)
	if err != nil {
		return false, err
	}

	if val2 == nil {
		return false, nil
	}

	ret, err := greaterEqual(val1, val2)
	if err != nil {
		return false, fmt.Errorf("%s: %w", path, err)
	}

	return ret, nil
}

func greaterEqual(str string, val any) (bool, error) {
	s, err := parse(str, val)
	if err != nil {
		return false, err
	}

	switch v := val.(type) {
	case int:
		return v >= s.(int), nil

	case int64:
		return v >= s.(int64), nil

	case uint:
		return v >= s.(uint), nil

	case uint64:
		return v >= s.(uint64), nil

	case float32:
		return v >= s.(float32), nil

	case float64:
		return v >= s.(float64), nil

	case string:
		return v >= s.(string), nil

	case bool:
		return v == true || v == s.(bool), nil

	case []int:
		return containsGreaterEqual(v, s.(int)), nil

	case []int64:
		return containsGreaterEqual(v, s.(int64)), nil

	case []uint:
		return containsGreaterEqual(v, s.(uint)), nil

	case []uint64:
		return containsGreaterEqual(v, s.(uint64)), nil

	case []float32:
		return containsGreaterEqual(v, s.(float32)), nil

	case []float64:
		return containsGreaterEqual(v, s.(float64)), nil

	case []string:
		return containsGreaterEqual(v, s.(string)), nil

	case []bool:
		for _, iter := range v {
			if iter == true || iter == s.(bool) {
				return true, nil
			}
		}

		return false, nil

	case time.Time:
		tm := s.(*timeVal)
		trunc := v.Truncate(tm.precision)
		return trunc.Equal(tm.time) || trunc.After(tm.time), nil

	case []time.Time:
		tm := s.(*timeVal)
		for _, iter := range v {
			trunc := iter.Truncate(tm.precision)
			if trunc.Equal(tm.time) || trunc.After(tm.time) {
				return true, nil
			}
		}

		return false, nil

	case civil.Date:
		return v == s.(civil.Date) || v.After(s.(civil.Date)), nil

	case []civil.Date:
		for _, iter := range v {
			if iter == s.(civil.Date) || iter.After(s.(civil.Date)) {
				return true, nil
			}
		}

		return false, nil

	default:
		return false, fmt.Errorf("unsupported struct type (%T)", v)
	}
}

func containsGreaterEqual[E constraints.Ordered](s []E, v E) bool {
	for _, iter := range s {
		if iter >= v {
			return true
		}
	}
	return false
}
