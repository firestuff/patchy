package patchy

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/firestuff/patchy/metadata"
	"github.com/firestuff/patchy/potency"
	"github.com/firestuff/patchy/store"
	"github.com/firestuff/patchy/storebus"
	"github.com/gorilla/mux"
)

type API struct {
	router   *mux.Router
	sb       *storebus.StoreBus
	potency  *potency.Potency
	registry map[string]*config
}

type Metadata = metadata.Metadata

func NewFileStoreAPI(root string) (*API, error) {
	return NewAPI(store.NewFileStore(root))
}

func NewAPI(st store.Storer) (*API, error) {
	api := &API{
		router:   mux.NewRouter(),
		sb:       storebus.NewStoreBus(st),
		registry: map[string]*config{},
	}

	api.potency = potency.NewPotency(st)
	api.router.Use(api.potency.Middleware)

	return api, nil
}

func Register[T any](api *API) {
	obj := new(T)
	t := reflect.TypeOf(*obj)
	RegisterName[T](api, strings.ToLower(t.Name()))
}

func RegisterName[T any](api *API, typeName string) {
	cfg := newConfig[T](typeName)
	api.registry[cfg.typeName] = cfg
	api.registerHandlers(fmt.Sprintf("/%s", cfg.typeName), cfg)

	obj := new(T)
	typ := reflect.TypeOf(*obj)

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)

		if f.Type.Kind() != reflect.Slice {
			continue
		}

		if f.Type.Elem().Kind() != reflect.String {
			continue
		}

		// TODO: Support f.Tag override
		fName := strings.ToLower(f.Name)

		if _, found := api.registry[fName]; !found {
			continue
		}
	}
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.router.ServeHTTP(w, r)
}

func (api *API) registerHandlers(base string, cfg *config) {
	api.router.HandleFunc(
		base,
		func(w http.ResponseWriter, r *http.Request) { api.streamList(cfg, w, r) },
	).
		Methods("GET").
		Headers("Accept", "text/event-stream")

	api.router.HandleFunc(
		base,
		func(w http.ResponseWriter, r *http.Request) { api.getList(cfg, w, r) },
	).
		Methods("GET")

	api.router.HandleFunc(
		base,
		func(w http.ResponseWriter, r *http.Request) { api.post(cfg, w, r) },
	).
		Methods("POST").
		Headers("Content-Type", "application/json")

	api.router.HandleFunc(
		fmt.Sprintf("%s/{id}", base),
		func(w http.ResponseWriter, r *http.Request) { api.put(cfg, w, r) },
	).
		Methods("PUT").
		Headers("Content-Type", "application/json")

	api.router.HandleFunc(
		fmt.Sprintf("%s/{id}", base),
		func(w http.ResponseWriter, r *http.Request) { api.patch(cfg, w, r) },
	).
		Methods("PATCH").
		Headers("Content-Type", "application/json")

	api.router.HandleFunc(
		fmt.Sprintf("%s/{id}", base),
		func(w http.ResponseWriter, r *http.Request) { api.delete(cfg, w, r) },
	).
		Methods("DELETE")

	api.router.HandleFunc(
		fmt.Sprintf("%s/{id}", base),
		func(w http.ResponseWriter, r *http.Request) { api.stream(cfg, w, r) },
	).
		Methods("GET").
		Headers("Accept", "text/event-stream")

	api.router.HandleFunc(
		fmt.Sprintf("%s/{id}", base),
		func(w http.ResponseWriter, r *http.Request) { api.get(cfg, w, r) },
	).
		Methods("GET")
}
