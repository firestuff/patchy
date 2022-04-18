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
		larger := float32(math.Max(math.Abs(v1), math.Abs(float64(v2))))
		epsilon := math.Nextafter32(larger, math.MaxFloat32) - larger
		return math.Abs(v1-float64(v2)) < float64(epsilon), nil

	case float64:
		v1, err := strconv.ParseFloat(val1, 64)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}
		larger := math.Max(math.Abs(v1), math.Abs(v2))
		epsilon := math.Nextafter(larger, math.MaxFloat64) - larger
		return math.Abs(v1-v2) < epsilon, nil

	case string:
		return val1 == v2, nil

	case bool:
		v1, err := strconv.ParseBool(val1)
		if err != nil {
			return false, fmt.Errorf("%s: %w", path, err)
		}
		return v1 == v2, nil

	case []string:
		return slices.Contains(v2, val1), nil

	case time.Time:
		for _, layout := range []string{
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05-07:00",
		} {
			v1, err := time.Parse(layout, val1)
			if err != nil {
				continue
			}

			return v1.Equal(v2), nil
		}

		v1, err := strconv.ParseInt(val1, 10, 64)
		if err != nil {
			return false, fmt.Errorf("%s: unknown time format", path)
		}

		// UNIX Seconds: 2969-05-03
		// UNIX Millis:  1971-01-01
		// Intended to give us a wide range of useful values in both schemes
		if v1 > 31536000000 {
			return v2.Equal(time.UnixMilli(v1)), nil
		} else {
			return v2.Equal(time.Unix(v1, 0)), nil
		}

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

func errorPath(prev []string, part string) string {
	if len(prev) == 0 {
		return part
	} else {
		return fmt.Sprintf("%s.%s", strings.Join(prev, "."), part)
	}
}
