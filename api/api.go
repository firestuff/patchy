package api

import "fmt"
import "net/http"
import "time"

import "github.com/google/uuid"
import "github.com/gorilla/mux"

import "github.com/firestuff/patchy/metadata"
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

type mayUpdate[T any] interface {
	MayUpdate(*T, *http.Request) error
}

type mayDelete interface {
	MayDelete(*http.Request) error
}

type mayRead interface {
	MayRead(*http.Request) error
}

func Register[T any](api *API, t string, factory func() *T) {
	cfg := &config{
		Factory: func() any { return factory() },
	}

	obj := factory()

	if _, has := any(obj).(mayCreate); has {
		cfg.MayCreate = func(obj any, r *http.Request) error {
			return obj.(mayCreate).MayCreate(r)
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
		func(w http.ResponseWriter, r *http.Request) { api.create(t, cfg, w, r) },
	).
		Methods("POST").
		Headers("Content-Type", "application/json")

	api.router.HandleFunc(
		fmt.Sprintf("/%s/{id}", t),
		func(w http.ResponseWriter, r *http.Request) { api.update(t, cfg, w, r) },
	).
		Methods("PATCH").
		Headers("Content-Type", "application/json")

	api.router.HandleFunc(
		fmt.Sprintf("/%s/{id}", t),
		func(w http.ResponseWriter, r *http.Request) { api.del(t, cfg, w, r) },
	).
		Methods("DELETE")

	api.router.HandleFunc(
		fmt.Sprintf("/%s/{id}", t),
		func(w http.ResponseWriter, r *http.Request) { api.stream(t, cfg, w, r) },
	).
		Methods("GET").
		Headers("Accept", "text/event-stream")

	api.router.HandleFunc(
		fmt.Sprintf("/%s/{id}", t),
		func(w http.ResponseWriter, r *http.Request) { api.read(t, cfg, w, r) },
	).
		Methods("GET")
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.router.ServeHTTP(w, r)
}

func (api *API) create(t string, cfg *config, w http.ResponseWriter, r *http.Request) {
	obj := cfg.Factory()

	err := readJson(r, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadata.GetMetadata(obj).Id = uuid.NewString()

	if cfg.MayCreate != nil {
		err = cfg.MayCreate(obj, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	err = api.sb.Write(t, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = writeJson(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *API) update(t string, cfg *config, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj := cfg.Factory()

	metadata.GetMetadata(obj).Id = vars["id"]

	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	err := api.sb.Read(t, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	ifMatch := r.Header.Get("If-Match")
	if ifMatch != "" {
		if ifMatch != fmt.Sprintf(`"%s"`, metadata.GetMetadata(obj).Sha256) {
			http.Error(w, fmt.Sprintf("If-Match mismatch"), http.StatusPreconditionFailed)
			return
		}
	}

	patch := cfg.Factory()

	err = readJson(r, patch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Metadata is immutable or server-owned
	metadata.ClearMetadata(patch)

	if cfg.MayUpdate != nil {
		err = cfg.MayUpdate(obj, patch, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	err = merge(obj, patch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = api.sb.Write(t, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = writeJson(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *API) del(t string, cfg *config, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj := cfg.Factory()

	metadata.GetMetadata(obj).Id = vars["id"]

	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	err := api.sb.Read(t, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if cfg.MayDelete != nil {
		err = cfg.MayDelete(obj, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	err = api.sb.Delete(t, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *API) stream(t string, cfg *config, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	_, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	obj := cfg.Factory()

	metadata.GetMetadata(obj).Id = vars["id"]

	cfg.mu.RLock()
	// THIS LOCK REQUIRES MANUAL UNLOCKING IN ALL BRANCHES

	err := api.sb.Read(t, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		cfg.mu.RUnlock()
		return
	}

	if cfg.MayRead != nil {
		err = cfg.MayRead(obj, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			cfg.mu.RUnlock()
			return
		}
	}

	err = writeUpdate(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		cfg.mu.RUnlock()
		return
	}

	closeChan := w.(http.CloseNotifier).CloseNotify()
	objChan := api.sb.Subscribe(t, obj)
	ticker := time.NewTicker(5 * time.Second)

	cfg.mu.RUnlock()

	connected := true
	for connected {
		select {

		case <-closeChan:
			connected = false

		case msg, ok := <-objChan:
			if ok {
				err = writeUpdate(w, msg)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				writeDelete(w)
				connected = false
			}

		case <-ticker.C:
			writeHeartbeat(w)

		}
	}
}

func (api *API) read(t string, cfg *config, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj := cfg.Factory()

	metadata.GetMetadata(obj).Id = vars["id"]

	err := api.sb.Read(t, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if cfg.MayRead != nil {
		err = cfg.MayRead(obj, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	err = writeJson(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
