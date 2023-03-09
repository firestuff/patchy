package patchy

import (
	"errors"
	"net/http"
	"time"

	"github.com/firestuff/patchy/jsrest"
)

var ErrStreamingNotSupported = errors.New("streaming not supported")

func (api *API) stream(cfg *config, id string, w http.ResponseWriter, r *http.Request) error {
	if _, ok := w.(http.Flusher); !ok {
		return jsrest.Errorf(jsrest.ErrBadRequest, "stream failed (%w)", ErrStreamingNotSupported)
	}

	ch, err := api.sb.ReadStream(cfg.typeName, id, cfg.factory)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	defer api.sb.CloseReadStream(cfg.typeName, id, ch)

	obj := <-ch
	if obj == nil {
		return jsrest.Errorf(jsrest.ErrNotFound, "%s", id)
	}

	obj, err = cfg.checkRead(obj, r)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrUnauthorized, "read check failed (%w)", err)
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	// TODO: Wrap another function so we can do streaming errors in return values

	err = writeEvent(w, "initial", obj)
	if err != nil {
		_ = writeEvent(w, "error", jsrest.ToJSONError(err))
		return nil
	}

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-r.Context().Done():
			return nil

		case msg, ok := <-ch:
			if ok {
				msg, err = cfg.checkRead(msg, r)
				if err != nil {
					_ = writeEvent(w, "error", jsrest.ToJSONError(err))
					return nil
				}

				err = writeEvent(w, "update", msg)
				if err != nil {
					_ = writeEvent(w, "error", jsrest.ToJSONError(err))
					return nil
				}
			} else {
				err = writeEvent(w, "delete", emptyEvent)
				if err != nil {
					_ = writeEvent(w, "error", jsrest.ToJSONError(err))
				}
				return nil
			}

		case <-ticker.C:
			err = writeEvent(w, "heartbeat", emptyEvent)
			if err != nil {
				_ = writeEvent(w, "error", jsrest.ToJSONError(err))
				return nil
			}
		}
	}
}
