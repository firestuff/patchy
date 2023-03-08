package patchy

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

var ErrUnknownStreamFormat = errors.New("unknown _stream format")

func (api *API) streamList(cfg *config, w http.ResponseWriter, r *http.Request) {
	if _, ok := w.(http.Flusher); !ok {
		jse := jsrest.FromError(ErrStreamingNotSupported, jsrest.StatusBadRequest)
		jse.Write(w)

		return
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		e := fmt.Errorf("failed to parse URL query: %w", err)
		jse := jsrest.FromError(e, jsrest.StatusBadRequest)
		jse.Write(w)

		return
	}

	stream := params.Get("_stream")
	params.Del("_stream")

	parsed, jse := parseListParams(params)
	if jse != nil {
		jse.Write(w)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	switch stream {
	case "":
		fallthrough
	case "full":
		api.streamListFull(cfg, w, r, parsed)

	case "diff":
		api.streamListDiff(cfg, w, r, parsed)

	default:
		e := fmt.Errorf("%s: %w", stream, ErrUnknownStreamFormat)
		jse := jsrest.FromError(e, jsrest.StatusBadRequest)
		jse.Write(w)
	}
}

func (api *API) streamListFull(cfg *config, w http.ResponseWriter, r *http.Request, params *listParams) {
	// TODO: Push jsrest.Error down
	ch, err := api.sb.ListStream(cfg.typeName, cfg.factory)
	if err != nil {
		e := fmt.Errorf("failed to read list: %w", err)
		jse := jsrest.FromError(e, jsrest.StatusInternalServerError)
		_ = writeEvent(w, "error", jse)

		return
	}

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-r.Context().Done():
			return

		case <-ticker.C:
			jse := writeEvent(w, "heartbeat", emptyEvent)
			if jse != nil {
				_ = writeEvent(w, "error", jse)
				return
			}

			continue

		case list := <-ch:
			list, err = filterList(cfg, r, params, list)
			if err != nil {
				e := fmt.Errorf("failed to filter list: %w", err)
				jse := jsrest.FromError(e, jsrest.StatusBadRequest)
				_ = writeEvent(w, "error", jse)

				return
			}

			jse := writeEvent(w, "list", list)
			if jse != nil {
				_ = writeEvent(w, "error", jse)
				return
			}
		}
	}
}

func (api *API) streamListDiff(cfg *config, w http.ResponseWriter, r *http.Request, params *listParams) {
	ch, err := api.sb.ListStream(cfg.typeName, cfg.factory)
	if err != nil {
		e := fmt.Errorf("failed to read list: %w", err)
		jse := jsrest.FromError(e, jsrest.StatusInternalServerError)
		_ = writeEvent(w, "error", jse)

		return
	}

	last := map[string]any{}

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			jse := writeEvent(w, "heartbeat", emptyEvent)
			if jse != nil {
				_ = writeEvent(w, "error", jse)
				return
			}

			continue

		case <-r.Context().Done():
			return

		case list := <-ch:
			list, err = filterList(cfg, r, params, list)
			if err != nil {
				e := fmt.Errorf("failed to filter list: %w", err)
				jse := jsrest.FromError(e, jsrest.StatusBadRequest)
				_ = writeEvent(w, "error", jse)

				return
			}

			cur := map[string]any{}

			for _, obj := range list {
				objMD := metadata.GetMetadata(obj)

				lastObj, found := last[objMD.ID]
				if found {
					lastMD := metadata.GetMetadata(lastObj)
					if objMD.ETag != lastMD.ETag {
						jse := writeEvent(w, "update", obj)
						if jse != nil {
							_ = writeEvent(w, "error", jse)
							return
						}
					}
				} else {
					jse := writeEvent(w, "add", obj)
					if jse != nil {
						_ = writeEvent(w, "error", jse)
						return
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

				jse := writeEvent(w, "remove", old)
				if jse != nil {
					_ = writeEvent(w, "error", jse)
					return
				}

				delete(last, id)
			}
		}
	}
}
