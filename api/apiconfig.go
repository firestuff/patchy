package api

import "fmt"
import "net/http"
import "sync"

type APIConfig struct {
	Factory func() (any, error)
	Update  func(any, any) error

	MayCreate func(any, *http.Request) error
	MayUpdate func(any, any, *http.Request) error
	MayDelete func(any, *http.Request) error
	MayRead   func(any, *http.Request) error

	mu sync.RWMutex
}

func (conf *APIConfig) validate() error {
	if conf.Factory == nil {
		return fmt.Errorf("APIConfig.Factory must be set")
	}

	if conf.Update == nil {
		return fmt.Errorf("APIConfig.Update must be set")
	}

	if conf.MayCreate == nil {
		return fmt.Errorf("APIConfig.MayCreate must be set")
	}

	if conf.MayUpdate == nil {
		return fmt.Errorf("APIConfig.MayUpdate must be set")
	}

	if conf.MayDelete == nil {
		return fmt.Errorf("APIConfig.MayDelete must be set")
	}

	if conf.MayRead == nil {
		return fmt.Errorf("APIConfig.MayRead must be set")
	}

	return nil
}
