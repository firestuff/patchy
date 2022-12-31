package patchy

import (
	"net/http"
	"net/url"

	"github.com/firestuff/patchy/jsrest"
)

func (api *API) getList(cfg *config, w http.ResponseWriter, r *http.Request) {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parsed, jse := parseListParams(params)
	if jse != nil {
		jse.Write(w)
		return
	}

	v, jse := api.list(cfg, r, parsed)
	if jse != nil {
		jse.Write(w)
		return
	}

	jse = jsrest.WriteList(w, <-v.Chan())
	if jse != nil {
		jse.Write(w)
		return
	}
}
