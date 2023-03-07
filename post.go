package patchy

import (
	"fmt"
	"net/http"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
	"github.com/google/uuid"
)

func (api *API) post(cfg *config, w http.ResponseWriter, r *http.Request) {
	obj := cfg.factory()

	jse := jsrest.Read(r, obj)
	if jse != nil {
		jse.Write(w)
		return
	}

	metadata.GetMetadata(obj).ID = uuid.NewString()

	obj, jse = cfg.checkWrite(obj, nil, r)
	if jse != nil {
		jse.Write(w)
		return
	}

	err := api.sb.Write(cfg.typeName, obj)
	if err != nil {
		e := fmt.Errorf("failed to write: %w", err)
		jse := jsrest.FromError(e, jsrest.StatusInternalServerError)
		jse.Write(w)

		return
	}

	obj, jse = cfg.checkRead(obj, r)
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
