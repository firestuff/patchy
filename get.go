package patchy

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (api *API) get(cfg *config, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	v, err := api.sb.Read(r.Context(), cfg.typeName, vars["id"], cfg.factory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	obj := <-v.Chan()

	if obj == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if cfg.mayRead != nil {
		err = cfg.mayRead(obj, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	err = writeJSON(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
