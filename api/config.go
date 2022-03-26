package api

import "net/http"
import "sync"

type config struct {
	Factory func() any

	MayCreate func(any, *http.Request) error
	MayUpdate func(any, any, *http.Request) error
	MayDelete func(any, *http.Request) error
	MayRead   func(any, *http.Request) error

	mu sync.RWMutex
}
