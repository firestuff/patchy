package api

import "fmt"
import "net/http"
import "time"

import "github.com/google/uuid"
import "github.com/gorilla/mux"

import "github.com/firestuff/patchy/metadata"
import "github.com/firestuff/patchy/storebus"

type API struct {
	router *mux.Router
	sb     *storebus.StoreBus
}

func NewAPI(root string) (*API, error) {
	return &API{
		router: mux.NewRouter(),
		sb:     storebus.NewStoreBus(root),
	}, nil
}

func (api *API) Register(t string, config *APIConfig) error {
	err := config.validate()
	if err != nil {
		return err
	}

	api.router.HandleFunc(
		fmt.Sprintf("/%s", t),
		func(w http.ResponseWriter, r *http.Request) { api.create(t, config, w, r) },
	).
		Methods("POST").
		Headers("Content-Type", "application/json")

	api.router.HandleFunc(
		fmt.Sprintf("/%s/{id}", t),
		func(w http.ResponseWriter, r *http.Request) { api.update(t, config, w, r) },
	).
		Methods("PATCH").
		Headers("Content-Type", "application/json")

	api.router.HandleFunc(
		fmt.Sprintf("/%s/{id}", t),
		func(w http.ResponseWriter, r *http.Request) { api.del(t, config, w, r) },
	).
		Methods("DELETE")

	api.router.HandleFunc(
		fmt.Sprintf("/%s/{id}", t),
		func(w http.ResponseWriter, r *http.Request) { api.stream(t, config, w, r) },
	).
		Methods("GET").
		Headers("Accept", "text/event-stream")

	api.router.HandleFunc(
		fmt.Sprintf("/%s/{id}", t),
		func(w http.ResponseWriter, r *http.Request) { api.read(t, config, w, r) },
	).
		Methods("GET")

	return nil
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.router.ServeHTTP(w, r)
}

func (api *API) create(t string, config *APIConfig, w http.ResponseWriter, r *http.Request) {
	obj, err := config.Factory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = readJson(r, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadata.GetMetadata(obj).Id = uuid.NewString()

	err = config.MayCreate(obj, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

func (api *API) update(t string, config *APIConfig, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := config.Factory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadata.GetMetadata(obj).Id = vars["id"]

	config.mu.Lock()
	defer config.mu.Unlock()

	err = api.sb.Read(t, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	patch, err := config.Factory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = readJson(r, patch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Metadata is immutable or server-owned
	metadata.ClearMetadata(patch)

	err = config.MayUpdate(obj, patch, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
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

func (api *API) del(t string, config *APIConfig, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := config.Factory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadata.GetMetadata(obj).Id = vars["id"]

	config.mu.Lock()
	defer config.mu.Unlock()

	err = api.sb.Read(t, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = config.MayDelete(obj, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = api.sb.Delete(t, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *API) stream(t string, config *APIConfig, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	_, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	obj, err := config.Factory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadata.GetMetadata(obj).Id = vars["id"]

	config.mu.RLock()
	// THIS LOCK REQUIRES MANUAL UNLOCKING IN ALL BRANCHES

	err = api.sb.Read(t, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		config.mu.RUnlock()
		return
	}

	err = config.MayRead(obj, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		config.mu.RUnlock()
		return
	}

	err = writeUpdate(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		config.mu.RUnlock()
		return
	}

	closeChan := w.(http.CloseNotifier).CloseNotify()
	objChan := api.sb.Subscribe(t, obj)
	ticker := time.NewTicker(5 * time.Second)

	config.mu.RUnlock()

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

func (api *API) read(t string, config *APIConfig, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := config.Factory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadata.GetMetadata(obj).Id = vars["id"]

	err = api.sb.Read(t, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = config.MayRead(obj, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = writeJson(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
