package api

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
	"github.com/firestuff/patchy/potency"
	"github.com/firestuff/patchy/store"
	"github.com/firestuff/patchy/storebus"
	"github.com/julienschmidt/httprouter"
	"github.com/timewasted/go-accept-headers"
)

type API struct {
	router      *httprouter.Router
	sb          *storebus.StoreBus
	potency     *potency.Potency
	registry    map[string]*config
	requestHook RequestHook
}

type RequestHook func(*http.Request, *API) (*http.Request, error)

type Metadata = metadata.Metadata

var ErrUnknownAcceptType = errors.New("unknown Accept type")

func NewFileStoreAPI(root string) (*API, error) {
	return NewAPI(store.NewFileStore(root))
}

func NewSQLiteAPI(dbname string) (*API, error) {
	st, err := store.NewSQLiteStore(dbname)
	if err != nil {
		return nil, err
	}

	return NewAPI(st)
}

func NewAPI(st store.Storer) (*API, error) {
	router := httprouter.New()

	api := &API{
		router:   router,
		sb:       storebus.NewStoreBus(st),
		potency:  potency.NewPotency(st, router),
		registry: map[string]*config{},
	}

	api.router.GET(
		"/_debug",
		func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) { api.handleDebug(w, r) },
	)

	return api, nil
}

func Register[T any](api *API) {
	RegisterName[T](api, objName(new(T)))
}

func RegisterName[T any](api *API, typeName string) {
	// TODO: Support nested types
	cfg := newConfig[T](typeName)
	api.registry[cfg.typeName] = cfg
	api.registerHandlers(fmt.Sprintf("/%s", cfg.typeName), cfg)
}

func (api *API) SetRequestHook(hook RequestHook) {
	api.requestHook = hook
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
	if api.requestHook != nil {
		var err error
		r, err = api.requestHook(r, api)
		if err != nil {
			err = jsrest.Errorf(jsrest.ErrInternalServerError, "auth callback failed (%w)", err)
			jsrest.WriteError(w, err)
			return
		}
	}

	api.potency.ServeHTTP(w, r)
}

func (api *API) Close() {
	api.sb.Close()
}

func (api *API) registerHandlers(base string, cfg *config) {
	api.router.GET(
		base,
		func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			api.wrapError(api.routeListGET, cfg, w, r)
		},
	)

	api.router.POST(
		base,
		func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			api.wrapError(api.post, cfg, w, r)
		},
	)

	single := fmt.Sprintf("%s/:id", base)

	api.router.PUT(
		single,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			api.wrapErrorID(api.put, cfg, ps[0].Value, w, r)
		},
	)

	api.router.PATCH(
		single,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			api.wrapErrorID(api.patch, cfg, ps[0].Value, w, r)
		},
	)

	api.router.DELETE(
		single,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			api.wrapErrorID(api.delete, cfg, ps[0].Value, w, r)
		},
	)

	api.router.GET(
		single,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			api.wrapErrorID(api.routeSingleGET, cfg, ps[0].Value, w, r)
		},
	)
}

func (api *API) routeListGET(cfg *config, w http.ResponseWriter, r *http.Request) error {
	ac, err := accept.Negotiate(r.Header.Get("Accept"), "application/json", "text/event-stream")
	if err != nil {
		return jsrest.Errorf(jsrest.ErrNotAcceptable, "Accept: %s (%w)", r.Header.Get("Accept"), ErrUnknownAcceptType)
	}

	switch ac {
	case "application/json":
		return api.getList(cfg, w, r)
	case "text/event-stream":
		return api.streamList(cfg, w, r)
	default:
		return jsrest.Errorf(jsrest.ErrNotAcceptable, "Accept: %s (%w)", r.Header.Get("Accept"), ErrUnknownAcceptType)
	}
}

func (api *API) routeSingleGET(cfg *config, id string, w http.ResponseWriter, r *http.Request) error {
	ac, err := accept.Negotiate(r.Header.Get("Accept"), "application/json", "text/event-stream")
	if err != nil {
		return jsrest.Errorf(jsrest.ErrNotAcceptable, "Accept: %s (%w)", r.Header.Get("Accept"), ErrUnknownAcceptType)
	}

	switch ac {
	case "application/json":
		return api.get(cfg, id, w, r)
	case "text/event-stream":
		return api.stream(cfg, id, w, r)
	default:
		return jsrest.Errorf(jsrest.ErrNotAcceptable, "Accept: %s (%w)", r.Header.Get("Accept"), ErrUnknownAcceptType)
	}
}

func (api *API) wrapError(cb func(*config, http.ResponseWriter, *http.Request) error, cfg *config, w http.ResponseWriter, r *http.Request) {
	err := cb(cfg, w, r)
	if err != nil {
		jsrest.WriteError(w, err)
	}
}

func (api *API) wrapErrorID(cb func(*config, string, http.ResponseWriter, *http.Request) error, cfg *config, id string, w http.ResponseWriter, r *http.Request) {
	err := cb(cfg, id, w, r)
	if err != nil {
		jsrest.WriteError(w, err)
	}
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

	enc.Encode(map[string]any{ //nolint:errcheck,errchkjson
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

func objName[T any](obj *T) string {
	return strings.ToLower(reflect.TypeOf(*obj).Name())
}
