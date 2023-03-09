package patchy

import (
	"net/http"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

func (api *API) put(cfg *config, id string, w http.ResponseWriter, r *http.Request) {
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

	replace := cfg.factory()

	err = jsrest.Read(r, replace)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "read request failed (%w)", err)
		jsrest.WriteError(w, err)
		return
	}

	// Metadata is immutable or server-owned
	metadata.ClearMetadata(replace)
	objMD := metadata.GetMetadata(obj)
	replaceMD := metadata.GetMetadata(replace)
	replaceMD.ID = id
	replaceMD.Generation = objMD.Generation + 1

	replace, err = cfg.checkWrite(replace, prev, r)
	if err != nil {
		jsrest.WriteError(w, err)
		return
	}

	err = api.sb.Write(cfg.typeName, replace)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "write failed: %s (%w)", id, err)
		jsrest.WriteError(w, err)
		return
	}

	replace, err = cfg.checkRead(replace, r)
	if err != nil {
		jsrest.WriteError(w, err)
		return
	}

	err = jsrest.Write(w, replace)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "write response failed (%w)", err)
		jsrest.WriteError(w, err)
		return
	}
}
