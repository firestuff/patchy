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

	parsed, err := parseListParams(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	v, err := api.list(cfg, r, parsed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jse := jsrest.WriteList(w, <-v.Chan())
	if jse != nil {
		jse.Write(w)
		return
	}
}
