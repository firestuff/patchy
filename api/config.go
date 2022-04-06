package api

import "net/http"
import "sync"

type config struct {
	factory func() any

	mayCreate  func(any, *http.Request) error
	mayReplace func(any, any, *http.Request) error
	mayUpdate  func(any, any, *http.Request) error
	mayDelete  func(any, *http.Request) error
	mayRead    func(any, *http.Request) error

	mu sync.RWMutex
}
