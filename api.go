package patchy

import "encoding/json"
import "fmt"
import "net/http"
import "sync"
import "time"

import "github.com/google/uuid"
import "github.com/gorilla/mux"

import "github.com/firestuff/patchy/metadata"

type API struct {
	router *mux.Router
	sb     *StoreBus
}

type APIConfig struct {
	Factory func() (interface{}, error)
	Update  func(interface{}, interface{}) error

	MayCreate func(interface{}, *http.Request) error
	MayUpdate func(interface{}, interface{}, *http.Request) error
	MayDelete func(interface{}, *http.Request) error
	MayRead   func(interface{}, *http.Request) error

	mu sync.RWMutex
}

func NewAPI(root string, configs map[string]*APIConfig) (*API, error) {
	api := &API{
		router: mux.NewRouter(),
		sb:     NewStoreBus(root),
	}

	for t, config := range configs {
		err := config.validate()
		if err != nil {
			return nil, err
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
	}

	return api, nil
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

	err = config.Update(obj, patch)
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

func readJson(r *http.Request, obj interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(obj)
}

func writeJson(w http.ResponseWriter, obj interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	return enc.Encode(obj)
}

func writeUpdate(w http.ResponseWriter, obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("Failed to encode JSON: %s", err)
	}

	fmt.Fprintf(w, "event: update\ndata: %s\n\n", data)
	w.(http.Flusher).Flush()

	return nil
}

func writeDelete(w http.ResponseWriter) {
	fmt.Fprintf(w, "event: delete\ndata: {}\n\n")
	w.(http.Flusher).Flush()
}

func writeHeartbeat(w http.ResponseWriter) {
	fmt.Fprintf(w, "event: heartbeat\ndata: {}\n\n")
	w.(http.Flusher).Flush()
}

func (conf *APIConfig) validate() error {
	if conf.Factory == nil {
		return fmt.Errorf("APIConfig.Factory must be set")
	}

	if conf.Update == nil {
		return fmt.Errorf("APIConfig.Update must be set")
	}

	if conf.MayCreate == nil {
		return fmt.Errorf("APIConfig.MayCreate must be set")
	}

	if conf.MayUpdate == nil {
		return fmt.Errorf("APIConfig.MayUpdate must be set")
	}

	if conf.MayDelete == nil {
		return fmt.Errorf("APIConfig.MayDelete must be set")
	}

	if conf.MayRead == nil {
		return fmt.Errorf("APIConfig.MayRead must be set")
	}

	return nil
}
