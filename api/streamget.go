package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

var ErrStreamingNotSupported = errors.New("streaming not supported")

func (api *API) streamGet(cfg *config, id string, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	if _, ok := w.(http.Flusher); !ok {
		return jsrest.Errorf(jsrest.ErrBadRequest, "stream failed (%w)", ErrStreamingNotSupported)
	}

	opts, err := parseGetOpts(r)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrBadRequest, "parse get parameters failed (%w)", err)
	}

	gsi, err := api.streamGetInt(ctx, cfg, id)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	defer gsi.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	err = api.streamGetWrite(ctx, w, gsi.ch, opts)
	if err != nil {
		_ = writeEvent(w, "error", "", jsrest.ToJSONError(err), true)
		return nil
	}

	return nil
}

func (api *API) streamGetWrite(ctx context.Context, w http.ResponseWriter, ch <-chan any, opts *GetOpts) error {
	eventType := "initial"
	ticker := time.NewTicker(5 * time.Second)

	initialETag := opts.IfNoneMatchETag
	initialGeneration := opts.IfNoneMatchGeneration

	for {
		select {
		case <-ctx.Done():
			return nil

		case obj, ok := <-ch:
			if !ok {
				err := writeEvent(w, "delete", "", emptyEvent, true)
				if err != nil {
					return jsrest.Errorf(jsrest.ErrInternalServerError, "write delete failed (%w)", err)
				}

				return nil
			}

			md := metadata.GetMetadata(obj)

			if initialETag != "" || initialGeneration > 0 {
				if md.ETag == initialETag || md.Generation == initialGeneration {
					err := writeEvent(w, "notModified", md.ETag, emptyEvent, true)
					if err != nil {
						return jsrest.Errorf(jsrest.ErrInternalServerError, "write update failed (%w)", err)
					}

					initialETag = ""
					initialGeneration = 0
					eventType = "update"

					continue
				}

				initialETag = ""
				initialGeneration = 0
			}

			err := writeEvent(w, eventType, md.ETag, obj, true)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "write update failed (%w)", err)
			}

			if eventType == "initial" {
				eventType = "update"
			}

		case <-ticker.C:
			err := writeEvent(w, "heartbeat", "", emptyEvent, true)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "write heartbeat failed (%w)", err)
			}
		}
	}
}
