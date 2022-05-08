package patchy

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (api *API) delete(cfg *config, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	v, err := api.sb.Read(cfg.typeName, vars["id"], cfg.factory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	obj := <-v.Chan()
	if obj == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if cfg.mayDelete != nil {
		err = cfg.mayDelete(obj, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	err = api.sb.Delete(cfg.typeName, vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
