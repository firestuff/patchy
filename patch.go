package patchy

import (
	"net/http"

	"github.com/firestuff/patchy/metadata"
	"github.com/gorilla/mux"
)

func (api *API) patch(cfg *config, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	cfg.mu.Lock()
	defer cfg.mu.Unlock()

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

	if !ifMatch(obj, w, r) {
		return
	}

	patch := cfg.factory()

	err = readJSON(r, patch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Metadata is immutable or server-owned
	metadata.ClearMetadata(patch)

	if cfg.mayUpdate != nil {
		err = cfg.mayUpdate(obj, patch, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	merge(obj, patch)

	metadata.GetMetadata(obj).Generation++

	err = api.sb.Write(cfg.typeName, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = writeJSON(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
