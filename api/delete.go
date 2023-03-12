package api

import (
	"net/http"

	"github.com/firestuff/patchy/jsrest"
)

func (api *API) delete(cfg *config, id string, w http.ResponseWriter, r *http.Request) error {
	obj, err := api.sb.Read(cfg.typeName, id, cfg.factory)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	if obj == nil {
		return jsrest.Errorf(jsrest.ErrNotFound, "%s", id)
	}

	_, err = cfg.checkWrite(nil, obj, r)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrUnauthorized, "write check failed (%w)", err)
	}

	err = api.sb.Delete(cfg.typeName, id)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "delete failed: %s (%w)", id, err)
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}
