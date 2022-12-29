package patchy

import (
	"net/http"

	"github.com/firestuff/patchy/metadata"
)

func (api *API) patch(cfg *config, id string, w http.ResponseWriter, r *http.Request) {
	// TODO: Parse semicolon params
	switch r.Header.Get("Content-Type") {
	case "":
		fallthrough
	case "application/json":
		break

	default:
		http.Error(w, "unknown Content-Type", http.StatusUnsupportedMediaType)
		return
	}

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
