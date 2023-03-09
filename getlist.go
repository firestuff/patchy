package patchy

import (
	"net/http"
	"net/url"

	"github.com/firestuff/patchy/jsrest"
)

func (api *API) getList(cfg *config, w http.ResponseWriter, r *http.Request) {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrBadRequest, "parse URL query failed (%w)", err)
		jsrest.WriteError(w, err)
		return
	}

	parsed, err := parseListParams(params)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrBadRequest, "parse list parameters failed (%w)", err)
		jsrest.WriteError(w, err)
		return
	}

	// TODO: Push jsrest.Error down into storebus
	list, err := api.sb.List(cfg.typeName, cfg.factory)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "read list failed (%w)", err)
		jsrest.WriteError(w, err)
		return
	}

	list, err = filterList(cfg, r, parsed, list)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "filter list failed (%w)", err)
		jsrest.WriteError(w, err)
		return
	}

	err = jsrest.WriteList(w, list)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "write list failed (%w)", err)
		jsrest.WriteError(w, err)
		return
	}
}
