package patchy

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/firestuff/patchy/jsrest"
)

var ErrNotFound = errors.New("not found")

func (api *API) get(cfg *config, id string, w http.ResponseWriter, r *http.Request) {
	obj, err := api.sb.Read(cfg.typeName, id, cfg.factory)
	if err != nil {
		e := fmt.Errorf("failed to read %s: %w", id, err)
		jse := jsrest.FromError(e, jsrest.StatusInternalServerError)
		jse.Write(w)

		return
	}

	if obj == nil {
		e := fmt.Errorf("%s: %w", id, ErrNotFound)
		jse := jsrest.FromError(e, jsrest.StatusNotFound)
		jse.Write(w)

		return
	}

	obj, jse := cfg.checkRead(obj, r)
	if jse != nil {
		jse.Write(w)
		return
	}

	jse = jsrest.Write(w, obj)
	if jse != nil {
		jse.Write(w)
		return
	}
}
