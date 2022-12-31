package patchy

import (
	"net/http"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
	"github.com/google/uuid"
)

func (api *API) post(cfg *config, w http.ResponseWriter, r *http.Request) {
	obj := cfg.factory()

	if !jsrest.Read(w, r, obj) {
		return
	}

	metadata.GetMetadata(obj).ID = uuid.NewString()

	if cfg.mayCreate != nil {
		err := cfg.mayCreate(obj, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	err := api.sb.Write(cfg.typeName, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = jsrest.Write(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
