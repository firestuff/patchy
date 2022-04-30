package patchy

import (
	"errors"
	"net/http"
	"os"

	"github.com/firestuff/patchy/metadata"
	"github.com/gorilla/mux"
)

func (api *API) put(cfg *config, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj := cfg.factory()

	objMD := metadata.GetMetadata(obj)
	objMD.Id = vars["id"]

	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	err := api.sb.Read(cfg.typeName, obj)
	if errors.Is(err, os.ErrNotExist) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !ifMatch(obj, w, r) {
		return
	}

	replace := cfg.factory()

	err = readJson(r, replace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Metadata is immutable or server-owned
	metadata.ClearMetadata(replace)
	replaceMD := metadata.GetMetadata(replace)
	replaceMD.Id = vars["id"]
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

	err = writeJson(w, replace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
