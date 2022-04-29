package path

import "fmt"
import "time"

import "cloud.google.com/go/civil"
import "golang.org/x/exp/constraints"

func Less(obj any, path string, val1 string) (bool, error) {
	val2, err := getAny(obj, path)
	if err != nil {
		return false, err
	}

	if val2 == nil {
		return false, nil
	}

	ret, err := less(val1, val2)
	if err != nil {
		return false, fmt.Errorf("%s: %w", path, err)
	}

	return ret, nil
}

func less(str string, val any) (bool, error) {
	s, err := parse(str, val)
	if err != nil {
		return false, err
	}

	switch v := val.(type) {
	case int:
		return v < s.(int), nil

	case int64:
		return v < s.(int64), nil

	case uint:
		return v < s.(uint), nil

	case uint64:
		return v < s.(uint64), nil

	case float32:
		return v < s.(float32), nil

	case float64:
		return v < s.(float64), nil

	case string:
		return v < s.(string), nil

	case bool:
		return v == false && s.(bool) == true, nil

	case []int:
		return containsLess(v, s.(int)), nil

	case []int64:
		return containsLess(v, s.(int64)), nil

	case []uint:
		return containsLess(v, s.(uint)), nil

	case []uint64:
		return containsLess(v, s.(uint64)), nil

	case []float32:
		return containsLess(v, s.(float32)), nil

	case []float64:
		return containsLess(v, s.(float64)), nil

	case []string:
		return containsLess(v, s.(string)), nil

	case []bool:
		if s.(bool) != true {
			return false, nil
		}

		for _, iter := range v {
			if iter == false {
				return true, nil
			}
		}

		return false, nil

	case time.Time:
		tm := s.(*timeVal)
		return v.Truncate(tm.precision).Before(tm.time), nil

	case []time.Time:
		tm := s.(*timeVal)
		for _, iter := range v {
			if iter.Truncate(tm.precision).Before(tm.time) {
				return true, nil
			}
		}

		return false, nil

	case civil.Date:
		return v.Before(s.(civil.Date)), nil

	case []civil.Date:
		for _, iter := range v {
			if iter.Before(s.(civil.Date)) {
				return true, nil
			}
		}

		return false, nil

	default:
		return false, fmt.Errorf("unsupported struct type (%T)", v)
	}
}

func containsLess[E constraints.Ordered](s []E, v E) bool {
	for _, iter := range s {
		if iter < v {
			return true
		}
	}
	return false
}
