package api

import (
	"net/http"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

func (api *API) put(cfg *config, id string, w http.ResponseWriter, r *http.Request) error {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	obj, err := api.sb.Read(r.Context(), cfg.typeName, id, cfg.factory)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	if obj == nil {
		return jsrest.Errorf(jsrest.ErrNotFound, "%s", id)
	}

	err = ifMatch(obj, r)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "match failed (%w)", err)
	}

	prev, err := cfg.clone(obj)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "clone failed (%w)", err)
	}

	replace := cfg.factory()

	err = jsrest.Read(r, replace)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read request failed (%w)", err)
	}

	// Metadata is immutable or server-owned
	metadata.ClearMetadata(replace)
	objMD := metadata.GetMetadata(obj)
	replaceMD := metadata.GetMetadata(replace)
	replaceMD.ID = id
	replaceMD.Generation = objMD.Generation + 1

	replace, err = cfg.checkWrite(replace, prev, r)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrUnauthorized, "write check failed (%w)", err)
	}

	err = api.sb.Write(r.Context(), cfg.typeName, replace)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "write failed: %s (%w)", id, err)
	}

	replace, err = cfg.checkRead(replace, r)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrUnauthorized, "read check failed (%w)", err)
	}

	err = jsrest.Write(w, replace)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "write response failed (%w)", err)
	}

	return nil
}
