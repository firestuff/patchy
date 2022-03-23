package storebus

import "encoding/json"
import "fmt"
import "net/http"
import "time"

import "github.com/google/uuid"
import "github.com/gorilla/mux"

type API struct {
	router  *mux.Router
	sb      *StoreBus
}

type APIConfig struct {
	Factory func() (Object, error)
	Update  func(Object, Object) error

	MayCreate func(Object, *http.Request) error
	MayUpdate func(Object, Object, *http.Request) error
	MayRead   func(Object, *http.Request) error
}

func NewAPI(root string, configs map[string]*APIConfig) (*API, error) {
	api := &API{
		router:  mux.NewRouter(),
		sb:      NewStoreBus(root),
	}

	for t, config := range configs {
		err := config.validate()
		if err != nil {
			return nil, err
		}

		api.router.HandleFunc(
			fmt.Sprintf("/%s", t),
			func(w http.ResponseWriter, r *http.Request) { api.create(config, w, r) },
		).
			Methods("POST").
			Headers("Content-Type", "application/json")

		api.router.HandleFunc(
			fmt.Sprintf("/%s/{id}", t),
			func(w http.ResponseWriter, r *http.Request) { api.update(config, w, r) },
		).
			Methods("PATCH").
			Headers("Content-Type", "application/json")

		api.router.HandleFunc(
			fmt.Sprintf("/%s/{id}", t),
			func(w http.ResponseWriter, r *http.Request) { api.stream(config, w, r) },
		).
			Methods("GET").
			Headers("Accept", "text/event-stream")

		api.router.HandleFunc(
			fmt.Sprintf("/%s/{id}", t),
			func(w http.ResponseWriter, r *http.Request) { api.read(config, w, r) },
		).
			Methods("GET")
	}

	return api, nil
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.router.ServeHTTP(w, r)
}

func (api *API) create(config *APIConfig, w http.ResponseWriter, r *http.Request) {
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

	obj.SetId(uuid.NewString())

	err = config.MayCreate(obj, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = api.sb.Write(obj)
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

func (api *API) update(config *APIConfig, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := config.Factory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	obj.SetId(vars["id"])

	err = api.sb.Read(obj)
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

	err = api.sb.Write(obj)
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

func (api *API) stream(config *APIConfig, w http.ResponseWriter, r *http.Request) {
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

	obj.SetId(vars["id"])

	err = api.sb.Read(obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = config.MayRead(obj, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = writeEvent(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Fix race if the object is written between Read() and Subscribe()

	closeChan := w.(http.CloseNotifier).CloseNotify()
	objChan := api.sb.Subscribe(obj)
	ticker := time.NewTicker(5 * time.Second)

	connected := true
	for connected {
		select {

		case <-closeChan:
			connected = false

		case msg := <-objChan:
			err = writeEvent(w, msg)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		case <-ticker.C:
			writeEvent(w, newHeartbeat())

		}
	}
}

func (api *API) read(config *APIConfig, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := config.Factory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	obj.SetId(vars["id"])

	err = api.sb.Read(obj)
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

func readJson(r *http.Request, obj Object) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(obj)
}

func writeJson(w http.ResponseWriter, obj Object) error {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	return enc.Encode(obj)
}

type heartbeat struct {
}

func newHeartbeat() *heartbeat {
	return &heartbeat{}
}

func (h *heartbeat) GetType() string {
	return "heartbeat"
}

func (h *heartbeat) GetId() string {
	return ""
}

func (h *heartbeat) SetId(id string) {
}

func writeEvent(w http.ResponseWriter, obj Object) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("Failed to encode JSON: %s", err)
	}

	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", obj.GetType(), data)
	w.(http.Flusher).Flush()

	return nil
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

	if conf.MayRead == nil {
		return fmt.Errorf("APIConfig.MayRead must be set")
	}

	return nil
}
