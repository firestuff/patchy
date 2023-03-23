package api

import (
	"net/http"
	"net/url"

	"github.com/firestuff/patchy/jsrest"
)

func (api *API) getList(cfg *config, w http.ResponseWriter, r *http.Request) error {
	// TODO: Support If-None-Match
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrBadRequest, "parse URL query failed (%w)", err)
	}

	opts, err := parseListOpts(params)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrBadRequest, "parse list parameters failed (%w)", err)
	}

	list, err := api.listInt(r.Context(), cfg, opts)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "list failed (%w)", err)
	}

	err = jsrest.WriteList(w, list)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "write list failed (%w)", err)
	}

	return nil
}
