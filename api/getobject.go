package api

import (
	"net/http"

	"github.com/firestuff/patchy/jsrest"
)

func (api *API) getObject(cfg *config, id string, w http.ResponseWriter, r *http.Request) error {
	obj, err := api.getInt(r.Context(), cfg, r, id)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "get failed (%w)", err)
	}

	if obj == nil {
		return jsrest.Errorf(jsrest.ErrNotFound, "%s", id)
	}

	err = jsrest.Write(w, obj)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "write response failed (%w)", err)
	}

	return nil
}
