package patchy

import (
	"net/http"
	"os"

	"github.com/firestuff/patchy/metadata"
	"github.com/gorilla/mux"
)

func (api *API) get(cfg *config, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj := cfg.factory()

	metadata.GetMetadata(obj).Id = vars["id"]

	err := api.sb.Read(cfg.typeName, obj)
	if err == os.ErrNotExist {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if cfg.mayRead != nil {
		err = cfg.mayRead(obj, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	err = writeJson(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
