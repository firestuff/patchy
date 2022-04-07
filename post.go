package api

import "net/http"

import "github.com/google/uuid"

import "github.com/firestuff/patchy/metadata"

func (api *API) post(t string, cfg *config, w http.ResponseWriter, r *http.Request) {
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

	err = api.sb.Write(t, obj)
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
