package api

import "net/http"
import "os"

import "github.com/gorilla/mux"

import "github.com/firestuff/patchy/metadata"

func (api *API) delete(t string, cfg *config, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj := cfg.Factory()

	metadata.GetMetadata(obj).Id = vars["id"]

	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	err := api.sb.Read(t, obj)
	if err == os.ErrNotExist {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if cfg.MayDelete != nil {
		err = cfg.MayDelete(obj, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	err = api.sb.Delete(t, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
