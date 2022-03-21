package storebus

import "encoding/json"
import "net/http"

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

func NewAPI(root string, config *APIConfig) *API {
	api := &API{
		router: mux.NewRouter(),
		sb:     NewStoreBus(root),
		config: config,
	}

	api.router.HandleFunc("/{type}", api.create).Methods("POST").Headers("Content-Type", "application/json")
	api.router.HandleFunc("/{type}/{id}", api.update).Methods("PATCH").Headers("Content-Type", "application/json")
	api.router.HandleFunc("/{type}/{id}", api.read).Methods("GET")

	return api
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
