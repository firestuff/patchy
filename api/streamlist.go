package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

func (api *API) streamList(cfg *config, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	if _, ok := w.(http.Flusher); !ok {
		return jsrest.Errorf(jsrest.ErrBadRequest, "stream failed (%w)", ErrStreamingNotSupported)
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrBadRequest, "parse URL query failed (%w)", err)
	}

	opts, err := parseListOpts(params)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrBadRequest, "parse list parameters failed (%w)", err)
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	switch opts.Stream {
	case "":
		fallthrough
	case "full":
		err = api.streamListFull(ctx, cfg, w, opts)
		if err != nil {
			_ = writeEvent(w, "error", jsrest.ToJSONError(err))
		}

		return nil

	case "diff":
		err = api.streamListDiff(ctx, cfg, w, opts)
		if err != nil {
			_ = writeEvent(w, "error", jsrest.ToJSONError(err))
		}

		return nil

	default:
		return jsrest.Errorf(jsrest.ErrBadRequest, "_stream=%s (%w)", opts.Stream, ErrInvalidStreamFormat)
	}
}

func (api *API) streamListFull(ctx context.Context, cfg *config, w http.ResponseWriter, opts *ListOpts) error {
	// TODO: Add query condition pushdown
	lsi, err := api.streamListInt(ctx, cfg, opts)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read list failed (%w)", err)
	}
	defer lsi.Close()

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			err = writeEvent(w, "heartbeat", emptyEvent)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "write heartbeat failed (%w)", err)
			}

		case list := <-lsi.Chan():
			err = writeEvent(w, "list", list)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "write list failed (%w)", err)
			}
		}
	}
}

func (api *API) streamListDiff(ctx context.Context, cfg *config, w http.ResponseWriter, opts *ListOpts) error {
	lsi, err := api.streamListInt(ctx, cfg, opts)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read list failed (%w)", err)
	}
	defer lsi.Close()

	last := map[string]any{}

	ticker := time.NewTicker(5 * time.Second)

	// Force initial bytes across the connection, since otherwise diff mode could wait forever
	fmt.Fprintf(w, ":\n")
	w.(http.Flusher).Flush()

	for {
		select {
		case <-ticker.C:
			err = writeEvent(w, "heartbeat", emptyEvent)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "write heartbeat failed (%w)", err)
			}

			continue

		case <-ctx.Done():
			return nil

		case list := <-lsi.Chan():
			// TODO: Hash list, compare against previous and client If-Match (or similar)
			// TODO: If we sent changes or this is the first round, send some barrier
			cur := map[string]any{}

			for _, obj := range list {
				objMD := metadata.GetMetadata(obj)

				lastObj, found := last[objMD.ID]
				if found {
					lastMD := metadata.GetMetadata(lastObj)
					if objMD.ETag != lastMD.ETag {
						err = writeEvent(w, "update", obj)
						if err != nil {
							return jsrest.Errorf(jsrest.ErrInternalServerError, "write update failed (%w)", err)
						}
					}
				} else {
					err = writeEvent(w, "add", obj)
					if err != nil {
						return jsrest.Errorf(jsrest.ErrInternalServerError, "write add failed (%w)", err)
					}
				}

				cur[objMD.ID] = obj
				last[objMD.ID] = obj
			}

			for id, old := range last {
				_, found := cur[id]
				if found {
					continue
				}

				err = writeEvent(w, "remove", old)
				if err != nil {
					return jsrest.Errorf(jsrest.ErrInternalServerError, "write remove failed (%w)", err)
				}

				delete(last, id)
			}
		}
	}
}
