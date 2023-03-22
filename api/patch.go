package api

import (
	"net/http"

	"github.com/firestuff/patchy/jsrest"
)

func (api *API) patch(cfg *config, id string, w http.ResponseWriter, r *http.Request) error {
	patch := cfg.factory()

	err := jsrest.Read(r, patch)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read request failed (%w)", err)
	}

	opts, err := parseUpdateOpts(r)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrBadRequest, "parse update options (%w)", err)
	}

	obj, err := api.updateInt(r.Context(), cfg, id, patch, opts)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "update failed (%w)", err)
	}

	err = jsrest.Write(w, obj)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "write response failed (%w)", err)
	}

	return nil
}
