package path

import "fmt"
import "strconv"
import "time"

import "golang.org/x/exp/slices"

type timeMatch struct {
	time      time.Time
	precision time.Duration
}

func match(str string, val any) (bool, error) {
	s, err := parse(str, val)
	if err != nil {
		return false, err
	}

	switch v := val.(type) {
	case int:
		return s.(int) == v, nil

	case int64:
		return s.(int64) == v, nil

	case uint:
		return s.(uint) == v, nil

	case uint64:
		return s.(uint64) == v, nil

	case float32:
		return s.(float32) == v, nil

	case float64:
		return s.(float64) == v, nil

	case string:
		return s.(string) == v, nil

	case bool:
		return s.(bool) == v, nil

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
		tm := s.(*timeMatch)
		return tm.time.Equal(v.Truncate(tm.precision)), nil

	case []time.Time:
		tm := s.(*timeMatch)
		for _, iter := range v {
			if tm.time.Equal(iter.Truncate(tm.precision)) {
				return true, nil
			}
		}

		return false, nil

	default:
		return false, fmt.Errorf("unsupported struct type (%T)", v)
	}
}

func parse(str string, t any) (any, error) {
	switch v := t.(type) {
	case int:
		return parseInt(str)

	case []int:
		return parseInt(str)

	case int64:
		return strconv.ParseInt(str, 10, 64)

	case []int64:
		return strconv.ParseInt(str, 10, 64)

	case uint:
		return parseUint(str)

	case []uint:
		return parseUint(str)

	case uint64:
		return strconv.ParseUint(str, 10, 64)

	case []uint64:
		return strconv.ParseUint(str, 10, 64)

	case float32:
		return parseFloat32(str)

	case []float32:
		return parseFloat32(str)

	case float64:
		return strconv.ParseFloat(str, 64)

	case []float64:
		return strconv.ParseFloat(str, 64)

	case string:
		return str, nil

	case []string:
		return str, nil

	case bool:
		return strconv.ParseBool(str)

	case []bool:
		return strconv.ParseBool(str)

	case time.Time:
		return parseTime(str)

	case []time.Time:
		return parseTime(str)

	default:
		return nil, fmt.Errorf("unsupported struct type (%T)", v)
	}
}

func parseInt(str string) (int, error) {
	val, err := strconv.ParseInt(str, 10, strconv.IntSize)
	return int(val), err
}

func parseUint(str string) (uint, error) {
	val, err := strconv.ParseUint(str, 10, strconv.IntSize)
	return uint(val), err
}

func parseFloat32(str string) (float32, error) {
	val, err := strconv.ParseFloat(str, 32)
	return float32(val), err
}

func parseTime(str string) (*timeMatch, error) {
	for _, layout := range []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
	} {
		tm, err := time.Parse(layout, str)
		if err != nil {
			continue
		}
		return &timeMatch{
			time:      tm,
			precision: 1 * time.Second,
		}, nil
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unknown time format")
	}

	// UNIX Seconds: 2969-05-03
	// UNIX Millis:  1971-01-01
	// Intended to give us a wide range of useful values in both schemes
	if i > 31536000000 {
		return &timeMatch{
			time:      time.UnixMilli(i),
			precision: 1 * time.Millisecond,
		}, nil
	} else {
		return &timeMatch{
			time:      time.Unix(i, 0),
			precision: 1 * time.Second,
		}, nil
	}
}
