package path

import "fmt"
import "math"
import "reflect"
import "strings"
import "strconv"
import "time"

import "golang.org/x/exp/slices"

func Match(obj any, path string, val1 string) (bool, error) {
	val2, err := getAny(obj, path)
	if err != nil {
		return false, err
	}

	switch v2 := val2.(type) {
	case int:
		v1, err := strconv.ParseInt(val1, 10, 64)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}
		return v1 == int64(v2), nil

	case int64:
		v1, err := strconv.ParseInt(val1, 10, 64)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}
		return v1 == v2, nil

	case uint:
		v1, err := strconv.ParseUint(val1, 10, 64)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}
		return v1 == uint64(v2), nil

	case uint64:
		v1, err := strconv.ParseUint(val1, 10, 64)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}
		return v1 == v2, nil

	case float32:
		v1, err := strconv.ParseFloat(val1, 64)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}
		return float32Equal(v1, float64(v2)), nil

	case float64:
		v1, err := strconv.ParseFloat(val1, 64)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}
		return float64Equal(v1, v2), nil

	case string:
		return val1 == v2, nil

	case bool:
		v1, err := strconv.ParseBool(val1)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}
		return v1 == v2, nil

	case []int:
		v1, err := strconv.ParseInt(val1, 10, 64)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}
		return slices.Contains(v2, int(v1)), nil

	case []int64:
		v1, err := strconv.ParseInt(val1, 10, 64)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}
		return slices.Contains(v2, v1), nil

	case []uint:
		v1, err := strconv.ParseUint(val1, 10, 64)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}
		return slices.Contains(v2, uint(v1)), nil

	case []uint64:
		v1, err := strconv.ParseUint(val1, 10, 64)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}
		return slices.Contains(v2, v1), nil

	case []float32:
		v1, err := strconv.ParseFloat(val1, 64)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}

		for _, iter := range v2 {
			if float32Equal(v1, float64(iter)) {
				return true, nil
			}
		}

		return false, nil

	case []float64:
		v1, err := strconv.ParseFloat(val1, 64)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}

		for _, iter := range v2 {
			if float64Equal(v1, iter) {
				return true, nil
			}
		}

		return false, nil

	case []string:
		return slices.Contains(v2, val1), nil

	case []bool:
		v1, err := strconv.ParseBool(val1)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}
		return slices.Contains(v2, v1), nil

	case time.Time:
		v1, err := parseTime(val1)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}
		return v1.Equal(v2), nil

	case []time.Time:
		v1, err := parseTime(val1)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}

		for _, iter := range v2 {
			if v1.Equal(iter) {
				return true, nil
			}
		}

		return false, nil

	default:
		return false, fmt.Errorf("%s: unsupported struct type (%T)", path, val2)
	}

}

func getAny(obj any, path string) (any, error) {
	parts := strings.Split(path, ".")
	v := reflect.ValueOf(obj)
	return getAnyRecursive(v, parts, []string{})
}

func getAnyRecursive(v reflect.Value, parts []string, prev []string) (any, error) {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			// Not an error, just a lack of a result
			return nil, nil
		}

		v = reflect.Indirect(v)
	}

	if len(parts) == 0 {
		return v.Interface(), nil
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s: not a struct", strings.Join(prev, "."))
	}

	part := parts[0]

	sub := v.FieldByNameFunc(func(name string) bool {
		return strings.ToLower(part) == strings.ToLower(name)
	})
	if !sub.IsValid() {
		return nil, fmt.Errorf("%s: invalid field name", errorPath(prev, part))
	}

	newPrev := []string{}
	copy(newPrev, prev)
	newPrev = append(newPrev, part)

	return getAnyRecursive(sub, parts[1:], newPrev)
}

func parseTime(str string) (time.Time, error) {
	for _, layout := range []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
	} {
		tm, err := time.Parse(layout, str)
		if err != nil {
			continue
		}
		return tm, nil
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("unknown time format")
	}

	// UNIX Seconds: 2969-05-03
	// UNIX Millis:  1971-01-01
	// Intended to give us a wide range of useful values in both schemes
	if i > 31536000000 {
		return time.UnixMilli(i), nil
	} else {
		return time.Unix(i, 0), nil
	}
}

func float32Equal(f1 float64, f2 float64) bool {
	smaller := float64(math.Nextafter32(float32(f1), float32(math.Inf(-1))))
	larger := float64(math.Nextafter32(float32(f1), float32(math.Inf(+1))))
	return f2 >= smaller && f2 <= larger
}

func float64Equal(f1 float64, f2 float64) bool {
	smaller := math.Nextafter(f1, math.Inf(-1))
	larger := math.Nextafter(f1, math.Inf(+1))
	return f2 >= smaller && f2 <= larger
}

func errorPath(prev []string, part string) string {
	if len(prev) == 0 {
		return part
	} else {
		return fmt.Sprintf("%s.%s", strings.Join(prev, "."), part)
	}
}
