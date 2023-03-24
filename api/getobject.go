package api

import (
	"net/http"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

func (api *API) getObject(cfg *config, id string, w http.ResponseWriter, r *http.Request) error {
	opts, err := parseGetOpts(r)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrBadRequest, "parse get options (%w)", err)
	}

	obj, err := api.getInt(r.Context(), cfg, id)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "get failed (%w)", err)
	}

	if obj == nil {
		return jsrest.Errorf(jsrest.ErrNotFound, "%s", id)
	}

	md := metadata.GetMetadata(obj)

	if (opts.IfNoneMatchETag != "" && opts.IfNoneMatchETag == md.ETag) ||
		(opts.IfNoneMatchGeneration > 0 && opts.IfNoneMatchGeneration == md.Generation) {
		w.WriteHeader(http.StatusNotModified)
		return nil
	}

	err = jsrest.Write(w, obj)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "write response failed (%w)", err)
	}

	return nil
}
