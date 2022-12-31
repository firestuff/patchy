package patchy

import (
	"fmt"
	"net/http"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

func (api *API) patch(cfg *config, id string, w http.ResponseWriter, r *http.Request) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

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

	if !ifMatch(obj, w, r) {
		return
	}

	patch := cfg.factory()

	jse := jsrest.Read(r, patch)
	if jse != nil {
		jse.Write(w)
		return
	}

	// Metadata is immutable or server-owned
	metadata.ClearMetadata(patch)

	if cfg.mayUpdate != nil {
		err = cfg.mayUpdate(obj, patch, r)
		if err != nil {
			e := fmt.Errorf("unauthorized %s: %w", id, err)
			jse := jsrest.FromError(e, jsrest.StatusUnauthorized)
			jse.Write(w)

			return
		}
	}

	merge(obj, patch)

	metadata.GetMetadata(obj).Generation++

	err = api.sb.Write(cfg.typeName, obj)
	if err != nil {
		e := fmt.Errorf("failed to write %s: %w", id, err)
		jse := jsrest.FromError(e, jsrest.StatusInternalServerError)
		jse.Write(w)

		return
	}

	jse = jsrest.Write(w, obj)
	if jse != nil {
		jse.Write(w)
		return
	}
}
