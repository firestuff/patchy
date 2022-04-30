package patchy

import (
	"net/http"

	"github.com/firestuff/patchy/metadata"
	"github.com/google/uuid"
)

func (api *API) post(cfg *config, w http.ResponseWriter, r *http.Request) {
	obj := cfg.factory()

	err := readJson(r, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadata.GetMetadata(obj).Id = uuid.NewString()

	if cfg.mayCreate != nil {
		err = cfg.mayCreate(obj, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	err = api.sb.Write(cfg.typeName, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = writeJson(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
