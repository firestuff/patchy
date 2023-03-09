package patchy

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
		return fmt.Errorf("%s: MayRead: %w", cfg.typeName, ErrMissingAuthCheck)
	}

	if cfg.mayWrite == nil {
		return fmt.Errorf("%s: MayWrite: %w", cfg.typeName, ErrMissingAuthCheck)
	}

	return nil
}

func (cfg *config) checkRead(obj any, r *http.Request) (any, *jsrest.Error) {
	ret, err := cfg.clone(obj)
	if err != nil {
		// TODO: Replace fmt.Errorf+jsrest.FromError instances with jsrest.Errorf
		e := fmt.Errorf("clone failed: %w", err)
		return nil, jsrest.FromError(e, jsrest.StatusInternalServerError)
	}

	if cfg.mayRead != nil {
		err := cfg.mayRead(ret, r)
		if err != nil {
			e := fmt.Errorf("unauthorized: %w", err)
			return nil, jsrest.FromError(e, jsrest.StatusUnauthorized)
		}
	}

	return ret, nil
}

func (cfg *config) checkWrite(obj, prev any, r *http.Request) (any, *jsrest.Error) {
	var ret any

	if obj != nil {
		var err *jsrest.Error

		ret, err = cfg.clone(obj)
		if err != nil {
			e := fmt.Errorf("clone failed: %w", err)
			return nil, jsrest.FromError(e, jsrest.StatusInternalServerError)
		}
	}

	if cfg.mayWrite != nil {
		err := cfg.mayWrite(ret, prev, r)
		if err != nil {
			e := fmt.Errorf("unauthorized: %w", err)
			return nil, jsrest.FromError(e, jsrest.StatusUnauthorized)
		}
	}

	return ret, nil
}

func (cfg *config) clone(src any) (any, *jsrest.Error) {
	js, err := json.Marshal(src)
	if err != nil {
		e := fmt.Errorf("json marshal: %w", err)
		return nil, jsrest.FromError(e, jsrest.StatusInternalServerError)
	}

	dst := cfg.factory()

	err = json.Unmarshal(js, dst)
	if err != nil {
		e := fmt.Errorf("json unmarshal: %w", err)
		return nil, jsrest.FromError(e, jsrest.StatusInternalServerError)
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
