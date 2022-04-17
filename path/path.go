package path

import "fmt"
import "reflect"
import "strings"
import "strconv"

func Match(obj any, path string, val1 string) (bool, error) {
	val2, err := getAny(obj, path)
	if err != nil {
		return false, err
	}

	switch v2 := val2.(type) {
	case int:
		v1, err := strconv.ParseInt(val1, 10, 64)
		if err != nil {
			return false, err
		}
		return v1 == int64(v2), nil

	case int64:
		v1, err := strconv.ParseInt(val1, 10, 64)
		if err != nil {
			return false, err
		}
		return v1 == v2, nil

	case uint:
		v1, err := strconv.ParseUint(val1, 10, 64)
		if err != nil {
			return false, err
		}
		return v1 == uint64(v2), nil

	case uint64:
		v1, err := strconv.ParseUint(val1, 10, 64)
		if err != nil {
			return false, err
		}
		return v1 == v2, nil

	case string:
		return val1 == v2, nil

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
