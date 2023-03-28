package api

import (
	"net/http"

	"github.com/firestuff/patchy/jsrest"
)

func (api *API) delete(cfg *config, id string, w http.ResponseWriter, r *http.Request) error {
	// TODO: Support If-Match
	err := api.deleteInt(r.Context(), cfg, id)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "delete failed (%w)", err)
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}
