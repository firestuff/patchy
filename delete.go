package patchy

import (
	"net/http"
)

func (api *API) delete(cfg *config, id string, w http.ResponseWriter, r *http.Request) {
	v, err := api.sb.Read(r.Context(), cfg.typeName, id, cfg.factory)
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

	err = api.sb.Delete(cfg.typeName, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
