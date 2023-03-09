package patchy

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

var ErrUnknownStreamFormat = errors.New("unknown _stream format")

func (api *API) streamList(cfg *config, w http.ResponseWriter, r *http.Request) {
	if _, ok := w.(http.Flusher); !ok {
		err := jsrest.Errorf(jsrest.ErrBadRequest, "stream failed (%w)", ErrStreamingNotSupported)
		jsrest.WriteError(w, err)
		return
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrBadRequest, "parse URL query failed (%w)", err)
		jsrest.WriteError(w, err)
		return
	}

	stream := params.Get("_stream")
	params.Del("_stream")

	parsed, err := parseListParams(params)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrBadRequest, "parse list parameters failed (%w)", err)
		jsrest.WriteError(w, err)
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
		err = jsrest.Errorf(jsrest.ErrBadRequest, "_stream=%s (%w)", stream, ErrUnknownStreamFormat)
		jsrest.WriteError(w, err)
	}
}

func (api *API) streamListFull(cfg *config, w http.ResponseWriter, r *http.Request, params *listParams) {
	// TODO: Push jsrest.Error down
	ch, err := api.sb.ListStream(cfg.typeName, cfg.factory)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "read list failed (%w)", err)
		_ = writeEvent(w, "error", jsrest.ToJSONError(err))
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
				_ = writeEvent(w, "error", jsrest.ToJSONError(err))
				return
			}

			continue

		case list := <-ch:
			list, err = filterList(cfg, r, params, list)
			if err != nil {
				_ = writeEvent(w, "error", jsrest.ToJSONError(err))
				return
			}

			err = writeEvent(w, "list", list)
			if err != nil {
				_ = writeEvent(w, "error", jsrest.ToJSONError(err))
				return
			}
		}
	}
}

func (api *API) streamListDiff(cfg *config, w http.ResponseWriter, r *http.Request, params *listParams) {
	ch, err := api.sb.ListStream(cfg.typeName, cfg.factory)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "read list failed (%w)", err)
		_ = writeEvent(w, "error", jsrest.ToJSONError(err))
		return
	}

	last := map[string]any{}

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			err = writeEvent(w, "heartbeat", emptyEvent)
			if err != nil {
				_ = writeEvent(w, "error", jsrest.ToJSONError(err))
				return
			}

			continue

		case <-r.Context().Done():
			return

		case list := <-ch:
			list, err := filterList(cfg, r, params, list)
			if err != nil {
				_ = writeEvent(w, "error", jsrest.ToJSONError(err))
				return
			}

			cur := map[string]any{}

			for _, obj := range list {
				objMD := metadata.GetMetadata(obj)

				lastObj, found := last[objMD.ID]
				if found {
					lastMD := metadata.GetMetadata(lastObj)
					if objMD.ETag != lastMD.ETag {
						err = writeEvent(w, "update", obj)
						if err != nil {
							_ = writeEvent(w, "error", jsrest.ToJSONError(err))
							return
						}
					}
				} else {
					err = writeEvent(w, "add", obj)
					if err != nil {
						_ = writeEvent(w, "error", jsrest.ToJSONError(err))
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
					_ = writeEvent(w, "error", jsrest.ToJSONError(err))
					return
				}

				delete(last, id)
			}
		}
	}
}
