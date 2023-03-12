package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

var ErrMissingAuthCheck = errors.New("missing auth check")

type config struct {
	typeName string

	factory func() any

	mayRead  func(any, *http.Request) error
	mayWrite func(any, any, *http.Request) error

	mu sync.Mutex
}

type mayRead interface {
	MayRead(*http.Request) error
}

type mayWrite[T any] interface {
	MayWrite(*T, *http.Request) error
}

func newConfig[T any](typeName string) *config {
	cfg := &config{
		typeName: typeName,
		factory:  func() any { return new(T) },
	}

	typ := cfg.factory()

	if !metadata.HasMetadata(typ) {
		panic("struct missing patchy.Metadata field")
	}

	if _, has := typ.(mayRead); has {
		cfg.mayRead = func(obj any, r *http.Request) error {
			obj = convert[T](obj)
			return obj.(mayRead).MayRead(r)
		}
	}

	if _, found := typ.(mayWrite[T]); found {
		cfg.mayWrite = func(obj any, prev any, r *http.Request) error {
			obj = convert[T](obj)
			return obj.(mayWrite[T]).MayWrite(convert[T](prev), r)
		}
	}

	return cfg
}

func (cfg *config) isSafe() error {
	if cfg.mayRead == nil {
		return fmt.Errorf("%s: MayRead (%w)", cfg.typeName, ErrMissingAuthCheck)
	}

	if cfg.mayWrite == nil {
		return fmt.Errorf("%s: MayWrite (%w)", cfg.typeName, ErrMissingAuthCheck)
	}

	return nil
}

func (cfg *config) checkRead(obj any, r *http.Request) (any, error) {
	ret, err := cfg.clone(obj)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "clone failed (%w)", err)
	}

	if cfg.mayRead != nil {
		err := cfg.mayRead(ret, r)
		if err != nil {
			return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "not authorized to read (%w)", err)
		}
	}

	return ret, nil
}

func (cfg *config) checkWrite(obj, prev any, r *http.Request) (any, error) {
	var ret any

	if obj != nil {
		var err error
		ret, err = cfg.clone(obj)
		if err != nil {
			return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "clone failed (%w)", err)
		}
	}

	if cfg.mayWrite != nil {
		err := cfg.mayWrite(ret, prev, r)
		if err != nil {
			return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "not authorized to write (%w)", err)
		}
	}

	return ret, nil
}

func (cfg *config) clone(src any) (any, error) {
	js, err := json.Marshal(src)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "JSON marshal (%w)", err)
	}

	dst := cfg.factory()

	err = json.Unmarshal(js, dst)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "JSON unmarhsal (%w)", err)
	}

	return dst, nil
}

func convert[T any](obj any) *T {
	// Like cast but supports untyped nil
	if obj == nil {
		return nil
	}

	return obj.(*T)
}
