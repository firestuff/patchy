package api

import (
	"net/http"

	"github.com/firestuff/patchy/jsrest"
)

func (api *API) get(cfg *config, id string, w http.ResponseWriter, r *http.Request) error {
	obj, err := api.sb.Read(r.Context(), cfg.typeName, id, cfg.factory)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	if obj == nil {
		return jsrest.Errorf(jsrest.ErrNotFound, "%s", id)
	}

	obj, err = cfg.checkRead(obj, r)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrUnauthorized, "read check failed (%w)", err)
	}

	err = jsrest.Write(w, obj)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "write response failed (%w)", err)
	}

	return nil
}
