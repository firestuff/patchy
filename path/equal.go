package path

import "fmt"
import "time"

import "cloud.google.com/go/civil"
import "golang.org/x/exp/slices"

func Equal(obj any, path string, val1 string) (bool, error) {
	val2, err := getAny(obj, path)
	if err != nil {
		return false, err
	}

	if val2 == nil {
		return false, nil
	}

	ret, err := equal(val1, val2)
	if err != nil {
		return false, fmt.Errorf("%s: %w", path, err)
	}

	return ret, nil
}

func equal(str string, val any) (bool, error) {
	s, err := parse(str, val)
	if err != nil {
		return false, err
	}

	switch v := val.(type) {
	case []int:
		return slices.Contains(v, s.(int)), nil

	case []int64:
		return slices.Contains(v, s.(int64)), nil

	case []uint:
		return slices.Contains(v, s.(uint)), nil

	case []uint64:
		return slices.Contains(v, s.(uint64)), nil

	case []float32:
		return slices.Contains(v, s.(float32)), nil

	case []float64:
		return slices.Contains(v, s.(float64)), nil

	case []string:
		return slices.Contains(v, s.(string)), nil

	case []bool:
		return slices.Contains(v, s.(bool)), nil

	case time.Time:
		tm := s.(*timeVal)
		return tm.time.Equal(v.Truncate(tm.precision)), nil

	case []time.Time:
		tm := s.(*timeVal)
		for _, iter := range v {
			if tm.time.Equal(iter.Truncate(tm.precision)) {
				return true, nil
			}
		}

		return false, nil

	case []civil.Date:
		return slices.Contains(v, s.(civil.Date)), nil

	default:
		return s == val, nil
	}
}
