package path

import (
	"errors"
	"reflect"
	"sort"
	"time"

	"cloud.google.com/go/civil"
	"github.com/gopatchy/jsrest"
)

func Sort(objs any, path string) error {
	as := newAnySlice(objs, path)
	sort.Stable(as)

	return as.err
}

func SortReverse(objs any, path string) error {
	as := newAnySlice(objs, path)
	sort.Stable(sort.Reverse(as))

	return as.err
}

type anySlice struct {
	path    string
	slice   reflect.Value
	swapper func(i, j int)
	err     error
}

var ErrUnsupportedSortType = errors.New("unsupported _sort type")

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
	v1, err := Get(as.slice.Index(i).Interface(), as.path)
	if err != nil {
		as.err = err
		// We have to obey the Less() contract even in error cases
		return i < j
	}

	v2, err := Get(as.slice.Index(j).Interface(), as.path)
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
		as.err = jsrest.Errorf(jsrest.ErrBadRequest, "%s: %T (%w)", as.path, t1, ErrUnsupportedSortType)
		return i < j
	}
}

func (as *anySlice) Swap(i, j int) {
	as.swapper(i, j)
}
