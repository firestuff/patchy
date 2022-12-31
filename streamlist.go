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

var (
	ErrStreamingNotSupported = errors.New("streaming not supported")
	ErrUnknownStreamFormat   = errors.New("unknown _stream format")
)

func (api *API) streamList(cfg *config, w http.ResponseWriter, r *http.Request) {
	if _, ok := w.(http.Flusher); !ok {
		jse := jsrest.FromError(ErrStreamingNotSupported, http.StatusBadRequest)
		jse.Write(w)

		return
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		e := fmt.Errorf("failed to parse URL query: %w", err)
		jse := jsrest.FromError(e, http.StatusBadRequest)
		jse.Write(w)

		return
	}

	stream := params.Get("_stream")
	params.Del("_stream")

	parsed, err := parseListParams(params)
	if err != nil {
		// TODO: Make parseListParams return *jsrest.Error
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		jse := jsrest.FromError(e, http.StatusBadRequest)
		jse.Write(w)
	}
}

func (api *API) streamListFull(cfg *config, w http.ResponseWriter, r *http.Request, params *listParams) {
	// TODO: Make api.list return *jsrest.Error
	v, err := api.list(cfg, r, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-r.Context().Done():
			return

		case <-ticker.C:
			err = writeEvent(w, "heartbeat", emptyEvent)
			if err != nil {
				return
			}

			continue

		case list := <-v.Chan():
			err = writeEvent(w, "list", list)
			if err != nil {
				return
			}
		}
	}
}

func (api *API) streamListDiff(cfg *config, w http.ResponseWriter, r *http.Request, params *listParams) {
	v, err := api.list(cfg, r, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	last := map[string]any{}

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			err = writeEvent(w, "heartbeat", emptyEvent)
			if err != nil {
				return
			}

			continue

		case <-r.Context().Done():
			return

		case list := <-v.Chan():
			cur := map[string]any{}

			for _, obj := range list {
				objMD := metadata.GetMetadata(obj)

				lastObj, found := last[objMD.ID]
				if found {
					lastMD := metadata.GetMetadata(lastObj)
					if objMD.ETag != lastMD.ETag {
						err = writeEvent(w, "update", obj)
						if err != nil {
							return
						}
					}
				} else {
					err = writeEvent(w, "add", obj)
					if err != nil {
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

				err = writeEvent(w, "remove", old)
				if err != nil {
					return
				}

				delete(last, id)
			}
		}
	}
}
