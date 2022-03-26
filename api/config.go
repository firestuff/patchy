package api

import "fmt"
import "net/http"
import "sync"

import "github.com/firestuff/patchy/metadata"

type config struct {
	Factory func() any

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

	obj := cfg.Factory()

	if !metadata.GetMetadataField(obj).IsValid() {
		return fmt.Errorf("Missing Metadata field")
	}

	return nil
}
