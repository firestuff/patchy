package patchy

import (
	"net/http"
	"net/url"
	"time"

	"github.com/firestuff/patchy/metadata"
)

func (api *API) streamList(cfg *config, w http.ResponseWriter, r *http.Request) {
	if _, ok := w.(http.Flusher); !ok {
		http.Error(w, "Streaming not supported", http.StatusBadRequest)
		return
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stream := params.Get("_stream")
	params.Del("_stream")

	parsed, err := parseListParams(params)
	if err != nil {
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
		http.Error(w, "Unknown _stream format", http.StatusBadRequest)
	}
}

func (api *API) streamListFull(cfg *config, w http.ResponseWriter, r *http.Request, params *listParams) {
	v, err := api.sb.List(r.Context(), cfg.typeName, cfg.factory)
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
			list2, err := filterList(cfg, r, params, list)
			if err != nil {
				return
			}

			err = writeEvent(w, "list", list2)
			if err != nil {
				return
			}
		}
	}
}

func (api *API) streamListDiff(cfg *config, w http.ResponseWriter, r *http.Request, params *listParams) {
	v, err := api.sb.List(r.Context(), cfg.typeName, cfg.factory)
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
			list2, err := filterList(cfg, r, params, list)
			if err != nil {
				return
			}

			cur := map[string]any{}

			for _, obj := range list2 {
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
