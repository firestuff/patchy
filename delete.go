package patchy

import (
	"errors"
	"net/http"
	"os"

	"github.com/firestuff/patchy/metadata"
	"github.com/gorilla/mux"
)

func (api *API) delete(cfg *config, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj := cfg.factory()

	metadata.GetMetadata(obj).Id = vars["id"]

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
