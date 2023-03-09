package patchy

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
)

type API struct {
	router   *httprouter.Router
	sb       *storebus.StoreBus
	potency  *potency.Potency
	registry map[string]*config
}

type Metadata = metadata.Metadata

var ErrUnknownAcceptType = errors.New("unknown Accept type")

func NewFileStoreAPI(root string) (*API, error) {
	return NewAPI(store.NewFileStore(root))
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
	api.potency.ServeHTTP(w, r)
}

func (api *API) registerHandlers(base string, cfg *config) {
	api.router.GET(
		base,
		func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) { api.routeListGET(cfg, w, r) },
	)

	api.router.POST(
		base,
		func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) { api.post(cfg, w, r) },
	)

	single := fmt.Sprintf("%s/:id", base)

	api.router.PUT(
		single,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) { api.put(cfg, ps[0].Value, w, r) },
	)

	api.router.PATCH(
		single,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) { api.patch(cfg, ps[0].Value, w, r) },
	)

	api.router.DELETE(
		single,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) { api.delete(cfg, ps[0].Value, w, r) },
	)

	api.router.GET(
		single,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			api.routeSingleGET(cfg, ps[0].Value, w, r)
		},
	)
}

func (api *API) routeListGET(cfg *config, w http.ResponseWriter, r *http.Request) {
	// TODO: Parse Accept preference lists
	switch r.Header.Get("Accept") {
	case "text/event-stream":
		api.streamList(cfg, w, r)

	case "":
		fallthrough
	case "*/*":
		fallthrough
	case "application/json":
		api.getList(cfg, w, r)

	default:
		err := jsrest.Errorf(jsrest.ErrNotAcceptable, "Accept: %s (%w)", r.Header.Get("Accept"), ErrUnknownAcceptType)
		jsrest.WriteError(w, err)
	}
}

func (api *API) routeSingleGET(cfg *config, id string, w http.ResponseWriter, r *http.Request) {
	// TODO: Parse Accept preference lists
	switch r.Header.Get("Accept") {
	case "text/event-stream":
		api.stream(cfg, id, w, r)

	case "":
		fallthrough
	case "*/*":
		fallthrough
	case "application/json":
		api.get(cfg, id, w, r)

	default:
		err := jsrest.Errorf(jsrest.ErrNotAcceptable, "Accept: %s (%w)", r.Header.Get("Accept"), ErrUnknownAcceptType)
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
