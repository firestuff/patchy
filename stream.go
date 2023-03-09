package patchy

import (
	"errors"
	"net/http"
	"time"

	"github.com/firestuff/patchy/jsrest"
)

var ErrStreamingNotSupported = errors.New("streaming not supported")

func (api *API) stream(cfg *config, id string, w http.ResponseWriter, r *http.Request) {
	if _, ok := w.(http.Flusher); !ok {
		err := jsrest.Errorf(jsrest.ErrBadRequest, "stream failed (%w)", ErrStreamingNotSupported)
		jsrest.WriteError(w, err)
		return
	}

	ch, err := api.sb.ReadStream(cfg.typeName, id, cfg.factory)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
		jsrest.WriteError(w, err)
		return
	}

	defer api.sb.CloseReadStream(cfg.typeName, id, ch)

	obj := <-ch
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

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	err = writeEvent(w, "initial", obj)
	if err != nil {
		_ = writeEvent(w, "error", jsrest.ToJSONError(err))
		return
	}

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-r.Context().Done():
			return

		case msg, ok := <-ch:
			if ok {
				msg, err = cfg.checkRead(msg, r)
				if err != nil {
					_ = writeEvent(w, "error", jsrest.ToJSONError(err))
					return
				}

				err = writeEvent(w, "update", msg)
				if err != nil {
					_ = writeEvent(w, "error", jsrest.ToJSONError(err))
					return
				}
			} else {
				err = writeEvent(w, "delete", emptyEvent)
				if err != nil {
					_ = writeEvent(w, "error", jsrest.ToJSONError(err))
				}
				return
			}

		case <-ticker.C:
			err = writeEvent(w, "heartbeat", emptyEvent)
			if err != nil {
				_ = writeEvent(w, "error", jsrest.ToJSONError(err))
				return
			}
		}
	}
}
