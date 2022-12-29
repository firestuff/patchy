package patchy

import (
	"net/http"

	"github.com/firestuff/patchy/metadata"
)

func (api *API) put(cfg *config, id string, w http.ResponseWriter, r *http.Request) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

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

	if !ifMatch(obj, w, r) {
		return
	}

	replace := cfg.factory()

	if !readJSON(w, r, replace) {
		return
	}

	// Metadata is immutable or server-owned
	metadata.ClearMetadata(replace)
	objMD := metadata.GetMetadata(obj)
	replaceMD := metadata.GetMetadata(replace)
	replaceMD.ID = id
	replaceMD.Generation = objMD.Generation + 1

	if cfg.mayReplace != nil {
		err = cfg.mayReplace(obj, replace, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	err = api.sb.Write(cfg.typeName, replace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = writeJSON(w, replace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
