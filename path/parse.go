package path

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"cloud.google.com/go/civil"
)

type timeVal struct {
	time      time.Time
	precision time.Duration
}

var errUnsupportedType = fmt.Errorf("unsupported type")

func parse(str string, t any) (any, error) {
	typ := reflect.TypeOf(t)

	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
	}

	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	switch typ.Kind() { //nolint: exhaustive
	case reflect.Int:
		return parseInt(str)

	case reflect.Int64:
		return strconv.ParseInt(str, 10, 64)

	case reflect.Uint:
		return parseUint(str)

	case reflect.Uint64:
		return strconv.ParseUint(str, 10, 64)

	case reflect.Float32:
		return parseFloat32(str)

	case reflect.Float64:
		return strconv.ParseFloat(str, 64)

	case reflect.String:
		return str, nil

	case reflect.Bool:
		return strconv.ParseBool(str)

	case reflect.Struct:
		switch typ {
		case reflect.TypeOf(time.Time{}):
			return parseTime(str)

		case reflect.TypeOf(civil.Date{}):
			return civil.ParseDate(str)
		}
	}

	return nil, fmt.Errorf("%T: %w", t, errUnsupportedType)
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
		return nil, fmt.Errorf("unknown time format: %w", err)
	}

	// UNIX Seconds: 2969-05-03
	// UNIX Millis:  1971-01-01
	// Intended to give us a wide range of useful values in both schemes
	if i > 31536000000 {
		return &timeVal{
			time:      time.UnixMilli(i),
			precision: 1 * time.Millisecond,
		}, nil
	}

	return &timeVal{
		time:      time.Unix(i, 0),
		precision: 1 * time.Second,
	}, nil
}
