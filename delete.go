package patchy

import (
	"net/http"

	"github.com/firestuff/patchy/jsrest"
)

func (api *API) delete(cfg *config, id string, w http.ResponseWriter, r *http.Request) {
	obj, err := api.sb.Read(cfg.typeName, id, cfg.factory)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
		jsrest.WriteError(w, err)
		return
	}

	if obj == nil {
		err := jsrest.Errorf(jsrest.ErrNotFound, "%s", id)
		jsrest.WriteError(w, err)
		return
	}

	_, err = cfg.checkWrite(nil, obj, r)
	if err != nil {
		jsrest.WriteError(w, err)
		return
	}

	err = api.sb.Delete(cfg.typeName, id)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "delete failed: %s (%w)", id, err)
		jsrest.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
