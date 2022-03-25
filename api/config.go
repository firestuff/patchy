package api

import "fmt"
import "net/http"
import "sync"

type ConfigGetter interface {
	Get() *config
}

type Config[T any] struct {
	Factory func() (*T, error)

	MayCreate func(*T, *http.Request) error
	MayUpdate func(*T, *T, *http.Request) error
	MayDelete func(*T, *http.Request) error
	MayRead   func(*T, *http.Request) error
}

func (cfg Config[T]) Get() *config {
	return &config{
		Factory:   func() (any, error) { return cfg.Factory() },
		MayCreate: func(obj any, r *http.Request) error { return cfg.MayCreate(obj.(*T), r) },
		MayUpdate: func(obj any, patch any, r *http.Request) error { return cfg.MayUpdate(obj.(*T), patch.(*T), r) },
		MayDelete: func(obj any, r *http.Request) error { return cfg.MayDelete(obj.(*T), r) },
		MayRead:   func(obj any, r *http.Request) error { return cfg.MayRead(obj.(*T), r) },
	}
}

type config struct {
	Factory func() (any, error)

	MayCreate func(any, *http.Request) error
	MayUpdate func(any, any, *http.Request) error
	MayDelete func(any, *http.Request) error
	MayRead   func(any, *http.Request) error

	mu sync.RWMutex
}

func (cfg *config) validate() error {
	if cfg.Factory == nil {
		return fmt.Errorf("APIConfig.Factory must be set")
	}

	if cfg.MayCreate == nil {
		return fmt.Errorf("APIConfig.MayCreate must be set")
	}

	if cfg.MayUpdate == nil {
		return fmt.Errorf("APIConfig.MayUpdate must be set")
	}

	if cfg.MayDelete == nil {
		return fmt.Errorf("APIConfig.MayDelete must be set")
	}

	if cfg.MayRead == nil {
		return fmt.Errorf("APIConfig.MayRead must be set")
	}

	return nil
}
