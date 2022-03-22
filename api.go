package storebus

import "encoding/json"
import "fmt"
import "net/http"
import "time"

import "github.com/google/uuid"
import "github.com/gorilla/mux"

type API struct {
	router *mux.Router
	sb     *StoreBus
	config *APIConfig
}

type APIConfig struct {
	Factory func(string) (Object, error)
	Update  func(Object, Object) error

	MayCreate func(Object, *http.Request) error
	MayUpdate func(Object, Object, *http.Request) error
	MayRead   func(Object, *http.Request) error
}

func NewAPI(root string, config *APIConfig) (*API, error) {
	err := config.validate()
	if err != nil {
		return nil, err
	}

	api := &API{
		router: mux.NewRouter(),
		sb:     NewStoreBus(root),
		config: config,
	}

	api.router.HandleFunc("/{type}", api.create).Methods("POST").Headers("Content-Type", "application/json")
	api.router.HandleFunc("/{type}/{id}", api.update).Methods("PATCH").Headers("Content-Type", "application/json")
	api.router.HandleFunc("/{type}/{id}", api.stream).Methods("GET").Headers("Accept", "text/event-stream")
	api.router.HandleFunc("/{type}/{id}", api.read).Methods("GET")

	return api, nil
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.router.ServeHTTP(w, r)
}

func (api *API) create(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := api.config.Factory(vars["type"])
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

	err = api.config.MayCreate(obj, r)
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

func (api *API) update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	oldObj, err := api.config.Factory(vars["type"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	oldObj.SetId(vars["id"])

	err = api.sb.Read(oldObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	newObj, err := api.config.Factory(vars["type"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = readJson(r, newObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = api.config.MayUpdate(oldObj, newObj, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = api.config.Update(oldObj, newObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = api.sb.Write(oldObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = writeJson(w, oldObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *API) stream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	_, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	obj, err := api.config.Factory(vars["type"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	obj.SetId(vars["id"])

	closeChan := w.(http.CloseNotifier).CloseNotify()
	objChan := api.sb.Subscribe(obj)
	ticker := time.NewTicker(5 * time.Second)

	connected := true
	first := true

	for connected {
		select {

		case <-closeChan:
			connected = false

		case msg := <-objChan:
			if first {
				first = false

				err = api.config.MayRead(msg, r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
			}

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

func (api *API) read(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := api.config.Factory(vars["type"])
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

	err = api.config.MayRead(obj, r)
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
