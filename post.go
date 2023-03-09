package patchy

import (
	"net/http"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
	"github.com/google/uuid"
)

func (api *API) post(cfg *config, w http.ResponseWriter, r *http.Request) {
	obj := cfg.factory()

	err := jsrest.Read(r, obj)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "read request failed (%w)", err)
		jsrest.WriteError(w, err)
		return
	}

	metadata.GetMetadata(obj).ID = uuid.NewString()

	obj, err = cfg.checkWrite(obj, nil, r)
	if err != nil {
		jsrest.WriteError(w, err)
		return
	}

	err = api.sb.Write(cfg.typeName, obj)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "write failed (%w)", err)
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
