package patchy

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/firestuff/patchy/jsrest"
)

var ErrStreamingNotSupported = errors.New("streaming not supported")

func (api *API) stream(cfg *config, id string, w http.ResponseWriter, r *http.Request) {
	if _, ok := w.(http.Flusher); !ok {
		jse := jsrest.FromError(ErrStreamingNotSupported, jsrest.StatusBadRequest)
		jse.Write(w)

		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	ch, err := api.sb.ReadStream(cfg.typeName, id, cfg.factory)
	if err != nil {
		e := fmt.Errorf("failed to read %s: %w", id, err)
		jse := jsrest.FromError(e, jsrest.StatusInternalServerError)
		jse.Write(w)

		return
	}

	defer api.sb.CloseReadStream(cfg.typeName, id, ch)

	obj := <-ch
	if obj == nil {
		e := fmt.Errorf("%s: %w", id, ErrNotFound)
		jse := jsrest.FromError(e, jsrest.StatusNotFound)
		jse.Write(w)

		return
	}

	obj, jse := cfg.checkRead(obj, r)
	if jse != nil {
		jse.Write(w)
		return
	}

	jse = writeEvent(w, "initial", obj)
	if jse != nil {
		_ = writeEvent(w, "error", jse)
		return
	}

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-r.Context().Done():
			return

		case msg, ok := <-ch:
			if ok {
				msg, jse = cfg.checkRead(msg, r)
				if jse != nil {
					jse.Write(w)
					return
				}

				jse = writeEvent(w, "update", msg)
				if jse != nil {
					_ = writeEvent(w, "error", jse)
					return
				}
			} else {
				jse = writeEvent(w, "delete", emptyEvent)
				if jse != nil {
					_ = writeEvent(w, "error", jse)
				}
				return
			}

		case <-ticker.C:
			jse = writeEvent(w, "heartbeat", emptyEvent)
			if jse != nil {
				_ = writeEvent(w, "error", jse)
				return
			}
		}
	}
}
