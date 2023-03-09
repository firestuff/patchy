package patchy

import (
	"net/http"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

func (api *API) patch(cfg *config, id string, w http.ResponseWriter, r *http.Request) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

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

	err = ifMatch(obj, r)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "match failed (%w)", err)
		jsrest.WriteError(w, err)
		return
	}

	prev, err := cfg.clone(obj)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "clone failed (%w)", err)
		jsrest.WriteError(w, err)
		return
	}

	patch := cfg.factory()

	err = jsrest.Read(r, patch)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "read request failed (%w)", err)
		jsrest.WriteError(w, err)
		return
	}

	// Metadata is immutable or server-owned
	metadata.ClearMetadata(patch)

	merge(obj, patch)

	metadata.GetMetadata(obj).Generation++

	obj, err = cfg.checkWrite(obj, prev, r)
	if err != nil {
		jsrest.WriteError(w, err)
		return
	}

	err = api.sb.Write(cfg.typeName, obj)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "write failed: %s (%w)", id, err)
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
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "write response failed (%w)", err)
		jsrest.WriteError(w, err)
		return
	}
}
