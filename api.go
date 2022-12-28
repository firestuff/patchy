package patchy

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

	api.router.HandleFunc(
		"/_debug",
		func(w http.ResponseWriter, r *http.Request) { api.handleDebug(w, r) },
	).
		Methods("GET")

	return api, nil
}

func Register[T any](api *API) {
	obj := new(T)
	t := reflect.TypeOf(*obj)
	RegisterName[T](api, strings.ToLower(t.Name()))
}

func RegisterName[T any](api *API, typeName string) {
	// TODO: Support nested types
	cfg := newConfig[T](typeName)
	api.registry[cfg.typeName] = cfg
	api.registerHandlers(fmt.Sprintf("/%s", cfg.typeName), cfg)
}

func (api *API) IsSafe() error {
	for _, cfg := range api.registry {
		err := cfg.isSafe()
		if err != nil {
			return err
		}
	}

	return nil
}

func (api *API) CheckSafe() {
	err := api.IsSafe()
	if err != nil {
		panic(err)
	}
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.router.ServeHTTP(w, r)
}

// TODO: Add standard HTTP error handling that returns JSON

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

func (api *API) handleDebug(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "\t")

	hostname, _ := os.Hostname()

	if r.TLS == nil {
		r.TLS = &tls.ConnectionState{}
	}

	enc.Encode(map[string]any{ //nolint: errcheck,errchkjson
		"server": map[string]any{
			"hostname": hostname,
		},
		"ip": map[string]any{
			"remoteaddr": r.RemoteAddr,
		},
		"http": map[string]any{
			"proto":  r.Proto,
			"method": r.Method,
			"header": r.Header,
			"url":    r.URL.String(),
		},
		"tls": map[string]any{
			"version":            r.TLS.Version,
			"didresume":          r.TLS.DidResume,
			"ciphersuite":        r.TLS.CipherSuite,
			"negotiatedprotocol": r.TLS.NegotiatedProtocol,
			"servername":         r.TLS.ServerName,
		},
	})
}
