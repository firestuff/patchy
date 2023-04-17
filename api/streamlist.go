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
	w.Header().Set("Stream-Format", opts.Stream)

	switch opts.Stream {
	case "full":
		err = api.streamListFull(ctx, cfg, w, opts)
		if err != nil {
			_ = writeEvent(w, "error", nil, jsrest.ToJSONError(err), true)
		}

		return nil

	case "diff":
		err = api.streamListDiff(ctx, cfg, w, opts)
		if err != nil {
			_ = writeEvent(w, "error", nil, jsrest.ToJSONError(err), true)
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
			err = writeEvent(w, "heartbeat", nil, emptyEvent, true)
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
					err = writeEvent(w, "notModified", map[string]string{"id": etag}, emptyEvent, true)
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

			err = writeEvent(w, "list", map[string]string{"id": etag}, list, true)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "write list failed (%w)", err)
			}
		}
	}
}

type listEntry struct {
	pos int
	obj any
}

func (api *API) streamListDiff(ctx context.Context, cfg *config, w http.ResponseWriter, opts *ListOpts) error {
	lsi, err := api.streamListInt(ctx, cfg, opts)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read list failed (%w)", err)
	}
	defer lsi.Close()

	last := map[string]*listEntry{}

	ticker := time.NewTicker(5 * time.Second)

	first := true
	previousETag := ""

	for {
		select {
		case <-ticker.C:
			err = writeEvent(w, "heartbeat", nil, emptyEvent, true)
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
				err = writeEvent(w, "notModified", map[string]string{"id": etag}, emptyEvent, true)
				if err != nil {
					return jsrest.Errorf(jsrest.ErrInternalServerError, "write list failed (%w)", err)
				}

				continue
			}

			if previousETag == etag {
				continue
			}

			previousETag = etag
			first = false

			cur := map[string]*listEntry{}

			for pos, obj := range list {
				objMD := metadata.GetMetadata(obj)

				lastEntry := last[objMD.ID]
				if lastEntry == nil {
					err = writeEvent(w, "add", nil, obj, false)
					if err != nil {
						return jsrest.Errorf(jsrest.ErrInternalServerError, "write add failed (%w)", err)
					}
				} else {
					lastMD := metadata.GetMetadata(lastEntry.obj)
					if objMD.ETag != lastMD.ETag {
						err = writeEvent(w, "update", nil, obj, false)
						if err != nil {
							return jsrest.Errorf(jsrest.ErrInternalServerError, "write update failed (%w)", err)
						}
					}
				}

				cur[objMD.ID] = &listEntry{
					pos: pos,
					obj: obj,
				}

				last[objMD.ID] = cur[objMD.ID]
			}

			for id, old := range last {
				if cur[id] != nil {
					continue
				}

				err = writeEvent(w, "remove", nil, old.obj, false)
				if err != nil {
					return jsrest.Errorf(jsrest.ErrInternalServerError, "write remove failed (%w)", err)
				}

				delete(last, id)
			}

			err = writeEvent(w, "sync", map[string]string{"id": etag}, emptyEvent, true)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "write sync failed (%w)", err)
			}
		}
	}
}
