package path

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"github.com/gopatchy/jsrest"
)

type timeVal struct {
	time      time.Time
	precision time.Duration
}

var (
	ErrUnsupportedType   = errors.New("unsupported type")
	ErrUnknownTimeFormat = errors.New("unknown time format")
)

func parse(str string, t any) (any, error) {
	typ := reflect.TypeOf(t)

	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
	}

	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	// TODO: Consider attempting to convert to string in default case
	switch typ.Kind() { //nolint:exhaustive
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

	return nil, jsrest.Errorf(jsrest.ErrBadRequest, "%T (%w)", t, ErrUnsupportedType)
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

type timeFormat struct {
	format    string
	precision time.Duration
}

var timeFormats = []timeFormat{
	{
		format:    "2006-01-02-07:00",
		precision: 24 * time.Hour,
		// TODO: Support field annotation to change start vs end of day
		// TODO: Support timezone context passed down to allow naked date
	},
	{
		format:    "2006-01-02T15:04:05Z",
		precision: 1 * time.Second,
	},
	{
		format:    "2006-01-02T15:04:05-07:00",
		precision: 1 * time.Second,
	},
}

func parseTime(str string) (*timeVal, error) {
	if strings.ToLower(str) == "now" {
		return &timeVal{
			time:      time.Now(),
			precision: 1 * time.Nanosecond,
		}, nil
	}

	for _, format := range timeFormats {
		tm, err := time.Parse(format.format, str)
		if err != nil {
			continue
		}

		return &timeVal{
			time:      tm,
			precision: format.precision,
		}, nil
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", str, ErrUnknownTimeFormat)
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
