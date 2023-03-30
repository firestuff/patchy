package api

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
	"github.com/firestuff/patchy/path"
	"github.com/firestuff/patchy/potency"
	"github.com/firestuff/patchy/store"
	"github.com/firestuff/patchy/storebus"
	"github.com/julienschmidt/httprouter"
	"github.com/vfaronov/httpheader"
)

type API struct {
	router   *httprouter.Router
	sb       *storebus.StoreBus
	potency  *potency.Potency
	registry map[string]*config

	openAPI openAPI

	stripPrefix RequestHook
	authBasic   RequestHook
	authBearer  RequestHook
	requestHook RequestHook
}

type (
	RequestHook func(*http.Request, *API) (*http.Request, error)
	ContextKey  int
	Metadata    = metadata.Metadata
)

var (
	ErrHeaderValueMissingQuotes = errors.New("header missing quotes")
	ErrUnknownAcceptType        = errors.New("unknown Accept type")
)

const (
	ContextInternal ContextKey = iota

	ContextAuthBearer
	ContextAuthBasic
)

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

	api.router.GET(
		"/_openapi",
		func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) { api.handleOpenAPI(w, r) },
	)

	api.router.ServeFiles(
		"/_swaggerui/*filepath",
		http.FS(swaggerUI),
	)

	api.registerTemplates()

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

	authBasicUserPath, ok := path.FindTagValueType(cfg.typeOf, "patchy", "authBasicUser")
	if ok {
		authBasicPassPath, ok := path.FindTagValueType(cfg.typeOf, "patchy", "authBasicPass")
		if !ok {
			panic("patchy:authBasicUser without patchy:authBasicPass")
		}

		SetAuthBasicName[T](api, typeName, authBasicUserPath, authBasicPassPath)
	}

	authBearerTokenPath, ok := path.FindTagValueType(cfg.typeOf, "patchy", "authBearerToken")
	if ok {
		SetAuthBearerName[T](api, typeName, authBearerTokenPath)
	}
}

func (api *API) SetStripPrefix(prefix string) {
	api.stripPrefix = func(r *http.Request, _ *API) (*http.Request, error) {
		if !strings.HasPrefix(r.URL.Path, prefix) {
			return nil, jsrest.Errorf(jsrest.ErrNotFound, "not found")
		}

		r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)

		return r, nil
	}
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
	var err error

	if api.stripPrefix != nil {
		r, err = api.stripPrefix(r, api)
		if err != nil {
			err = jsrest.Errorf(jsrest.ErrUnauthorized, "strip prefix failed (%w)", err)
			jsrest.WriteError(w, err)

			return
		}
	}

	if api.authBasic != nil {
		r, err = api.authBasic(r, api)
		if err != nil {
			err = jsrest.Errorf(jsrest.ErrUnauthorized, "basic authentication failed (%w)", err)
			jsrest.WriteError(w, err)

			return
		}
	}

	if api.authBearer != nil {
		r, err = api.authBearer(r, api)
		if err != nil {
			err = jsrest.Errorf(jsrest.ErrUnauthorized, "bearer authentication failed (%w)", err)
			jsrest.WriteError(w, err)

			return
		}
	}

	if api.requestHook != nil {
		var err error

		r, err = api.requestHook(r, api)
		if err != nil {
			err = jsrest.Errorf(jsrest.ErrInternalServerError, "request hook failed (%w)", err)
			jsrest.WriteError(w, err)

			return
		}
	}

	// TODO: Gate CORs with some kind of flag
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-store")

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
	ac := httpheader.Accept(r.Header)

	if m := httpheader.MatchAccept(ac, "application/json"); m.Type != "" {
		return api.getList(cfg, w, r)
	}

	if m := httpheader.MatchAccept(ac, "text/event-stream"); m.Type != "" {
		return api.streamList(cfg, w, r)
	}

	return jsrest.Errorf(jsrest.ErrNotAcceptable, "Accept: %s (%w)", r.Header.Get("Accept"), ErrUnknownAcceptType)
}

func (api *API) routeSingleGET(cfg *config, id string, w http.ResponseWriter, r *http.Request) error {
	ac := httpheader.Accept(r.Header)

	if m := httpheader.MatchAccept(ac, "application/json"); m.Type != "" {
		return api.getObject(cfg, id, w, r)
	}

	if m := httpheader.MatchAccept(ac, "text/event-stream"); m.Type != "" {
		return api.streamGet(cfg, id, w, r)
	}

	return jsrest.Errorf(jsrest.ErrNotAcceptable, "Accept: %s (%w)", r.Header.Get("Accept"), ErrUnknownAcceptType)
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

func (api *API) names() []string {
	names := []string{}
	for name := range api.registry {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

func objName[T any](obj *T) string {
	return strings.ToLower(reflect.TypeOf(*obj).Name())
}
