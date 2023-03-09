package patchy

import (
	"net/http"

	"github.com/firestuff/patchy/jsrest"
)

func (api *API) get(cfg *config, id string, w http.ResponseWriter, r *http.Request) {
	obj, err := api.sb.Read(cfg.typeName, id, cfg.factory)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
		jsrest.WriteError(w, err)
		return
	}

	if obj == nil {
		err = jsrest.Errorf(jsrest.ErrNotFound, "%s", id)
		jsrest.WriteError(w, err)
		return
	}

	obj, err = cfg.checkRead(obj, r)
	if err != nil {
		jsrest.WriteError(w, err)
		return
	}

	err = jsrest.Write(w, obj)
	if err != nil {
		jsrest.WriteError(w, err)
		return
	}
}
