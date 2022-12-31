package patchy

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/firestuff/patchy/metadata"
)

var ErrMissingAuthCheck = errors.New("missing auth check")

type config struct {
	typeName string

	factory func() any

	mayCreate  func(any, *http.Request) error
	mayReplace func(any, any, *http.Request) error
	mayUpdate  func(any, any, *http.Request) error
	mayDelete  func(any, *http.Request) error
	mayRead    func(any, *http.Request) error

	mu sync.Mutex
}

type mayCreate interface {
	MayCreate(*http.Request) error
}

type mayReplace[T any] interface {
	MayReplace(*T, *http.Request) error
}

type mayUpdate[T any] interface {
	MayUpdate(*T, *http.Request) error
}

type mayDelete interface {
	MayDelete(*http.Request) error
}

type mayRead interface {
	MayRead(*http.Request) error
}

func newConfig[T any](typeName string) *config {
	cfg := &config{
		typeName: typeName,
		factory:  func() any { return new(T) },
	}

	obj := new(T)

	if !metadata.HasMetadata(obj) {
		panic("struct missing patchy.Metadata field")
	}

	if _, has := any(obj).(mayCreate); has {
		cfg.mayCreate = func(obj any, r *http.Request) error {
			return obj.(mayCreate).MayCreate(r)
		}
	}

	if _, found := any(obj).(mayReplace[T]); found {
		cfg.mayReplace = func(obj any, replace any, r *http.Request) error {
			return obj.(mayReplace[T]).MayReplace(replace.(*T), r)
		}
	}

	if _, found := any(obj).(mayUpdate[T]); found {
		cfg.mayUpdate = func(obj any, patch any, r *http.Request) error {
			return obj.(mayUpdate[T]).MayUpdate(patch.(*T), r)
		}
	}

	if _, has := any(obj).(mayDelete); has {
		cfg.mayDelete = func(obj any, r *http.Request) error {
			return obj.(mayDelete).MayDelete(r)
		}
	}

	if _, has := any(obj).(mayRead); has {
		cfg.mayRead = func(obj any, r *http.Request) error {
			return obj.(mayRead).MayRead(r)
		}
	}

	return cfg
}

func (cfg *config) isSafe() error {
	if cfg.mayCreate == nil {
		return fmt.Errorf("%s: MayCreate: %w", cfg.typeName, ErrMissingAuthCheck)
	}

	if cfg.mayReplace == nil {
		return fmt.Errorf("%s: MayReplace: %w", cfg.typeName, ErrMissingAuthCheck)
	}

	if cfg.mayUpdate == nil {
		return fmt.Errorf("%s: MayUpdate: %w", cfg.typeName, ErrMissingAuthCheck)
	}

	if cfg.mayDelete == nil {
		return fmt.Errorf("%s: MayDelete: %w", cfg.typeName, ErrMissingAuthCheck)
	}

	if cfg.mayRead == nil {
		return fmt.Errorf("%s: MayRead: %w", cfg.typeName, ErrMissingAuthCheck)
	}

	return nil
}
