package patchy

import (
	"net/http"

	"github.com/firestuff/patchy/jsrest"
)

func (api *API) get(cfg *config, id string, w http.ResponseWriter, r *http.Request) {
	v, err := api.sb.Read(r.Context(), cfg.typeName, id, cfg.factory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	obj := <-v.Chan()

	if obj == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if cfg.mayRead != nil {
		err = cfg.mayRead(obj, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	jse := jsrest.Write(w, obj)
	if jse != nil {
		jse.Write(w)
		return
	}
}
