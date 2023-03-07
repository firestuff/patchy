package patchy

import (
	"fmt"
	"net/http"

	"github.com/firestuff/patchy/jsrest"
)

func (api *API) delete(cfg *config, id string, w http.ResponseWriter, r *http.Request) {
	v, err := api.sb.Read(r.Context(), cfg.typeName, id, cfg.factory)
	if err != nil {
		e := fmt.Errorf("failed to read %s: %w", id, err)
		jse := jsrest.FromError(e, jsrest.StatusInternalServerError)
		jse.Write(w)

		return
	}

	obj := <-v.Chan()
	if obj == nil {
		e := fmt.Errorf("%s: %w", id, ErrNotFound)
		jse := jsrest.FromError(e, jsrest.StatusNotFound)
		jse.Write(w)

		return
	}

	_, jse := cfg.checkWrite(nil, obj, r)
	if jse != nil {
		jse.Write(w)
		return
	}

	err = api.sb.Delete(cfg.typeName, id)
	if err != nil {
		e := fmt.Errorf("failed to delete %s: %w", id, err)
		jse := jsrest.FromError(e, jsrest.StatusInternalServerError)
		jse.Write(w)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
