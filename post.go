package patchy

import (
	"net/http"

	"github.com/firestuff/patchy/metadata"
	"github.com/google/uuid"
)

func (api *API) post(cfg *config, w http.ResponseWriter, r *http.Request) {
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

	obj := cfg.factory()

	err := readJSON(r, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadata.GetMetadata(obj).ID = uuid.NewString()

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

	err = writeJSON(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
