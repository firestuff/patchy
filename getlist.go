package patchy

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/firestuff/patchy/jsrest"
)

func (api *API) getList(cfg *config, w http.ResponseWriter, r *http.Request) {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		e := fmt.Errorf("failed to parse URL query: %w", err)
		jse := jsrest.FromError(e, jsrest.StatusBadRequest)
		jse.Write(w)

		return
	}

	parsed, jse := parseListParams(params)
	if jse != nil {
		jse.Write(w)
		return
	}

	list, jse := api.list(cfg, r, parsed)
	if jse != nil {
		jse.Write(w)
		return
	}

	// TODO: Filter list through checkRead()

	jse = jsrest.WriteList(w, list)
	if jse != nil {
		jse.Write(w)
		return
	}
}
