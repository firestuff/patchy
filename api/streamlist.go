package api

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

var ErrUnknownStreamFormat = errors.New("unknown _stream format")

func (api *API) streamList(cfg *config, w http.ResponseWriter, r *http.Request) error {
	if _, ok := w.(http.Flusher); !ok {
		return jsrest.Errorf(jsrest.ErrBadRequest, "stream failed (%w)", ErrStreamingNotSupported)
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrBadRequest, "parse URL query failed (%w)", err)
	}

	stream := params.Get("_stream")
	params.Del("_stream")

	parsed, err := parseListParams(params)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrBadRequest, "parse list parameters failed (%w)", err)
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	switch stream {
	case "":
		fallthrough
	case "full":
		err = api.streamListFull(cfg, w, r, parsed)
		if err != nil {
			writeEvent(w, "error", jsrest.ToJSONError(err))
		}
		return nil

	case "diff":
		err = api.streamListDiff(cfg, w, r, parsed)
		if err != nil {
			writeEvent(w, "error", jsrest.ToJSONError(err))
		}
		return nil

	default:
		return jsrest.Errorf(jsrest.ErrBadRequest, "_stream=%s (%w)", stream, ErrUnknownStreamFormat)
	}
}

func (api *API) streamListFull(cfg *config, w http.ResponseWriter, r *http.Request, params *listParams) error {
	// TODO: Add query condition pushdown

	ch, err := api.sb.ListStream(cfg.typeName, cfg.factory)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read list failed (%w)", err)
	}
	defer api.sb.CloseListStream(cfg.typeName, ch)

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-r.Context().Done():
			return nil

		case <-ticker.C:
			err = writeEvent(w, "heartbeat", emptyEvent)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "write heartbeat failed (%w)", err)
			}

		case list := <-ch:
			list, err = filterList(cfg, r, params, list)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "filter list failed (%w)", err)
			}

			err = writeEvent(w, "list", list)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "write list failed (%w)", err)
			}
		}
	}
}

func (api *API) streamListDiff(cfg *config, w http.ResponseWriter, r *http.Request, params *listParams) error {
	// TODO: Add query condition pushdown

	ch, err := api.sb.ListStream(cfg.typeName, cfg.factory)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read list failed (%w)", err)
	}
	defer api.sb.CloseListStream(cfg.typeName, ch)

	last := map[string]any{}

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			err = writeEvent(w, "heartbeat", emptyEvent)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "write heartbeat failed (%w)", err)
			}

			continue

		case <-r.Context().Done():
			return nil

		case list := <-ch:
			list, err := filterList(cfg, r, params, list)
			if err != nil {
				return jsrest.Errorf(jsrest.ErrInternalServerError, "filter list failed (%w)", err)
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
