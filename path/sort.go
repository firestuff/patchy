package path

import "fmt"
import "reflect"
import "sort"
import "time"

import "cloud.google.com/go/civil"

func Sort(objs any, path string) error {
	as := newAnySlice(objs, path)
	sort.Sort(as)
	return as.Error()
}

func SortReverse(objs any, path string) error {
	as := newAnySlice(objs, path)
	sort.Sort(sort.Reverse(as))
	return as.Error()
}

type anySlice struct {
	path    string
	slice   reflect.Value
	swapper func(i, j int)
	err     error
}

func newAnySlice(objs any, path string) *anySlice {
	return &anySlice{
		path:    path,
		slice:   reflect.ValueOf(objs),
		swapper: reflect.Swapper(objs),
	}
}

func (as *anySlice) Len() int {
	return as.slice.Len()
}

func (as *anySlice) Less(i, j int) bool {
	v1, err := getAny(as.slice.Index(i).Interface(), as.path)
	if err != nil {
		as.err = err
		// We have to obey the Less() contract even in error cases
		return i < j
	}

	v2, err := getAny(as.slice.Index(j).Interface(), as.path)
	if err != nil {
		as.err = err
		return i < j
	}

	switch {
	case v1 == nil && v2 == nil:
		return false
	case v1 == nil:
		return true
	case v2 == nil:
		return false
	}

	switch t1 := v1.(type) {
	case int:
		return t1 < v2.(int)

	case int64:
		return t1 < v2.(int64)

	case uint:
		return t1 < v2.(uint)

	case uint64:
		return t1 < v2.(uint64)

	case float32:
		return t1 < v2.(float32)

	case float64:
		return t1 < v2.(float64)

	case string:
		return t1 < v2.(string)

	case bool:
		return !t1 && v2.(bool)

	case time.Time:
		return t1.Before(v2.(time.Time))

	case civil.Date:
		return t1.Before(v2.(civil.Date))

	default:
		as.err = fmt.Errorf("unsupported sort type (%T)", t1)
		return i < j
	}
}

func (as *anySlice) Swap(i, j int) {
	as.swapper(i, j)
}

func (as *anySlice) Error() error {
	if as.err == nil {
		return nil
	} else {
		return fmt.Errorf("%s: %w", as.path, as.err)
	}
}
