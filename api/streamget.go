package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/firestuff/patchy/jsrest"
)

var ErrStreamingNotSupported = errors.New("streaming not supported")

func (api *API) streamGet(cfg *config, id string, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	if _, ok := w.(http.Flusher); !ok {
		return jsrest.Errorf(jsrest.ErrBadRequest, "stream failed (%w)", ErrStreamingNotSupported)
	}

	gsi, err := api.streamGetInt(ctx, cfg, id)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	defer gsi.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	err = api.streamGetWrite(ctx, w, gsi.ch)
	if err != nil {
		_ = writeEvent(w, "error", jsrest.ToJSONError(err), true)
		return nil
	}

	return nil
}

func (api *API) streamGetWrite(ctx context.Context, w http.ResponseWriter, ch <-chan any) error {
	eventType := "initial"

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil

		case msg, ok := <-ch:
			if ok {
				err := writeEvent(w, eventType, msg, true)
				if err != nil {
					return jsrest.Errorf(jsrest.ErrInternalServerError, "write update failed (%w)", err)
				}

				if eventType == "initial" {
					eventType = "update"
				}
			} else {
				err := writeEvent(w, "delete", emptyEvent, true)
				if err != nil {
					return jsrest.Errorf(jsrest.ErrInternalServerError, "write delete failed (%w)", err)
				}
				return nil
			}

		case <-ticker.C:
			err := writeEvent(w, "heartbeat", emptyEvent, true)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "write heartbeat failed (%w)", err)
			}
		}
	}
}
