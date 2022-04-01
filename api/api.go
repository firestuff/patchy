package api

import "fmt"
import "net/http"

import "github.com/gorilla/mux"

import "github.com/firestuff/patchy/potency"
import "github.com/firestuff/patchy/storebus"

type API struct {
	router  *mux.Router
	sb      *storebus.StoreBus
	potency *potency.Potency
}

func NewAPI(root string) (*API, error) {
	api := &API{
		router: mux.NewRouter(),
		sb:     storebus.NewStoreBus(root),
	}

	api.potency = potency.NewPotency(api.sb.GetStore())
	api.router.Use(api.potency.Middleware)

	return api, nil
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

func Register[T any](api *API, t string) {
	cfg := &config{
		Factory: func() any { return new(T) },
	}

	obj := new(T)

	if _, has := any(obj).(mayCreate); has {
		cfg.MayCreate = func(obj any, r *http.Request) error {
			return obj.(mayCreate).MayCreate(r)
		}
	}

	if _, found := any(obj).(mayReplace[T]); found {
		cfg.MayReplace = func(obj any, replace any, r *http.Request) error {
			return obj.(mayReplace[T]).MayReplace(replace.(*T), r)
		}
	}

	if _, found := any(obj).(mayUpdate[T]); found {
		cfg.MayUpdate = func(obj any, patch any, r *http.Request) error {
			return obj.(mayUpdate[T]).MayUpdate(patch.(*T), r)
		}
	}

	if _, has := any(obj).(mayDelete); has {
		cfg.MayDelete = func(obj any, r *http.Request) error {
			return obj.(mayDelete).MayDelete(r)
		}
	}

	if _, has := any(obj).(mayRead); has {
		cfg.MayRead = func(obj any, r *http.Request) error {
			return obj.(mayRead).MayRead(r)
		}
	}

	api.router.HandleFunc(
		fmt.Sprintf("/%s", t),
		func(w http.ResponseWriter, r *http.Request) { api.post(t, cfg, w, r) },
	).
		Methods("POST").
		Headers("Content-Type", "application/json")

	api.router.HandleFunc(
		fmt.Sprintf("/%s/{id}", t),
		func(w http.ResponseWriter, r *http.Request) { api.put(t, cfg, w, r) },
	).
		Methods("PUT").
		Headers("Content-Type", "application/json")

	api.router.HandleFunc(
		fmt.Sprintf("/%s/{id}", t),
		func(w http.ResponseWriter, r *http.Request) { api.patch(t, cfg, w, r) },
	).
		Methods("PATCH").
		Headers("Content-Type", "application/json")

	api.router.HandleFunc(
		fmt.Sprintf("/%s/{id}", t),
		func(w http.ResponseWriter, r *http.Request) { api.delete(t, cfg, w, r) },
	).
		Methods("DELETE")

	api.router.HandleFunc(
		fmt.Sprintf("/%s/{id}", t),
		func(w http.ResponseWriter, r *http.Request) { api.getStream(t, cfg, w, r) },
	).
		Methods("GET").
		Headers("Accept", "text/event-stream")

	api.router.HandleFunc(
		fmt.Sprintf("/%s/{id}", t),
		func(w http.ResponseWriter, r *http.Request) { api.get(t, cfg, w, r) },
	).
		Methods("GET")
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.router.ServeHTTP(w, r)
}
