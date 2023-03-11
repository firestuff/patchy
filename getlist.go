package patchy

import (
	"net/http"
	"net/url"

	"github.com/firestuff/patchy/jsrest"
)

func (api *API) getList(cfg *config, w http.ResponseWriter, r *http.Request) error {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrBadRequest, "parse URL query failed (%w)", err)
	}

	parsed, err := parseListParams(params)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrBadRequest, "parse list parameters failed (%w)", err)
	}

	// TODO: Add query condition pushdown

	list, err := api.sb.List(cfg.typeName, cfg.factory)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read list failed (%w)", err)
	}

	list, err = filterList(cfg, r, parsed, list)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "filter list failed (%w)", err)
	}

	err = jsrest.WriteList(w, list)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "write list failed (%w)", err)
	}

	return nil
}
