package path

import "fmt"
import "strconv"
import "time"

import "cloud.google.com/go/civil"

type timeVal struct {
	time      time.Time
	precision time.Duration
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

	case civil.Date:
		return civil.ParseDate(str)

	case []civil.Date:
		return civil.ParseDate(str)

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

func parseTime(str string) (*timeVal, error) {
	for _, layout := range []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
	} {
		tm, err := time.Parse(layout, str)
		if err != nil {
			continue
		}
		return &timeVal{
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
		return &timeVal{
			time:      time.UnixMilli(i),
			precision: 1 * time.Millisecond,
		}, nil
	} else {
		return &timeVal{
			time:      time.Unix(i, 0),
			precision: 1 * time.Second,
		}, nil
	}
}
