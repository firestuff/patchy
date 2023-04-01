package api

import (
	"context"
	"net/http"
	"time"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
	"github.com/vfaronov/httpheader"
)

func (api *API) streamList(cfg *config, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	if _, ok := w.(http.Flusher); !ok {
		return jsrest.Errorf(jsrest.ErrBadRequest, "stream failed (%w)", ErrStreamingNotSupported)
	}

	opts, err := api.parseListOpts(r)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrBadRequest, "parse list parameters failed (%w)", err)
	}

	w.Header().Set("Content-Type", "text/event-stream")

	switch opts.Stream {
	case "":
		fallthrough
	case "full":
		err = api.streamListFull(ctx, cfg, w, opts)
		if err != nil {
			_ = writeEvent(w, "error", "", jsrest.ToJSONError(err), true)
		}

		return nil

	case "diff":
		err = api.streamListDiff(ctx, cfg, w, opts)
		if err != nil {
			_ = writeEvent(w, "error", "", jsrest.ToJSONError(err), true)
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
	first := true
	previousETag := ""

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			err = writeEvent(w, "heartbeat", "", emptyEvent, true)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "write heartbeat failed (%w)", err)
			}

		case list := <-lsi.Chan():
			etag, err := HashList(list)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "hash list failed (%w)", err)
			}

			if first {
				first = false

				if httpheader.MatchWeak(opts.IfNoneMatch, httpheader.EntityTag{Opaque: etag}) {
					err = writeEvent(w, "notModified", etag, emptyEvent, true)
					if err != nil {
						return jsrest.Errorf(jsrest.ErrInternalServerError, "write list failed (%w)", err)
					}

					continue
				}
			}

			if previousETag == etag {
				continue
			}

			previousETag = etag

			err = writeEvent(w, "list", etag, list, true)
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

	first := true
	previousETag := ""

	for {
		select {
		case <-ticker.C:
			err = writeEvent(w, "heartbeat", "", emptyEvent, true)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "write heartbeat failed (%w)", err)
			}

			continue

		case <-ctx.Done():
			return nil

		case list := <-lsi.Chan():
			etag, err := HashList(list)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "hash list failed (%w)", err)
			}

			if first && httpheader.MatchWeak(opts.IfNoneMatch, httpheader.EntityTag{Opaque: etag}) {
				err = writeEvent(w, "notModified", etag, emptyEvent, true)
				if err != nil {
					return jsrest.Errorf(jsrest.ErrInternalServerError, "write list failed (%w)", err)
				}

				continue
			}

			if previousETag == etag {
				continue
			}

			previousETag = etag

			cur := map[string]any{}
			changed := false

			for _, obj := range list {
				objMD := metadata.GetMetadata(obj)

				lastObj, found := last[objMD.ID]
				if found {
					lastMD := metadata.GetMetadata(lastObj)
					if objMD.ETag != lastMD.ETag {
						changed = true

						err = writeEvent(w, "update", "", obj, false)
						if err != nil {
							return jsrest.Errorf(jsrest.ErrInternalServerError, "write update failed (%w)", err)
						}
					}
				} else {
					changed = true

					err = writeEvent(w, "add", "", obj, false)
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

				changed = true

				err = writeEvent(w, "remove", "", old, false)
				if err != nil {
					return jsrest.Errorf(jsrest.ErrInternalServerError, "write remove failed (%w)", err)
				}

				delete(last, id)
			}

			if first || changed {
				first = false

				err = writeEvent(w, "sync", etag, emptyEvent, true)
				if err != nil {
					return jsrest.Errorf(jsrest.ErrInternalServerError, "write sync failed (%w)", err)
				}
			}
		}
	}
}
