package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/firestuff/patchy/jsrest"
)

var ErrStreamingNotSupported = errors.New("streaming not supported")

func (api *API) stream(cfg *config, id string, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	if _, ok := w.(http.Flusher); !ok {
		return jsrest.Errorf(jsrest.ErrBadRequest, "stream failed (%w)", ErrStreamingNotSupported)
	}

	ch, err := api.sb.ReadStream(r.Context(), cfg.typeName, id, cfg.factory)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	defer api.sb.CloseReadStream(cfg.typeName, id, ch)

	obj := <-ch
	if obj == nil {
		return jsrest.Errorf(jsrest.ErrNotFound, "%s", id)
	}

	obj, err = cfg.checkRead(ctx, obj, api)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrUnauthorized, "read check failed (%w)", err)
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	err = api.streamInt(ctx, cfg, w, obj, ch)
	if err != nil {
		writeEvent(w, "error", jsrest.ToJSONError(err))
		return nil
	}

	return nil
}

func (api *API) streamInt(ctx context.Context, cfg *config, w http.ResponseWriter, obj any, ch <-chan any) error {
	err := writeEvent(w, "initial", obj)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "write initial failed (%w)", err)
	}

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil

		case msg, ok := <-ch:
			if ok {
				msg, err = cfg.checkRead(ctx, msg, api)
				if err != nil {
					return jsrest.Errorf(jsrest.ErrUnauthorized, "read check failed (%w)", err)
				}

				err = writeEvent(w, "update", msg)
				if err != nil {
					return jsrest.Errorf(jsrest.ErrInternalServerError, "write update failed (%w)", err)
				}
			} else {
				err = writeEvent(w, "delete", emptyEvent)
				if err != nil {
					return jsrest.Errorf(jsrest.ErrInternalServerError, "write delete failed (%w)", err)
				}
				return nil
			}

		case <-ticker.C:
			err = writeEvent(w, "heartbeat", emptyEvent)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "write heartbeat failed (%w)", err)
			}
		}
	}
}
