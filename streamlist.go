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
		api.streamListFull(cfg, w, r, params, parsed)

	case "diff":
		api.streamListDiff(cfg, w, r, params, parsed)

	default:
		http.Error(w, "Unknown _stream format", http.StatusBadRequest)
	}
}

func (api *API) streamListFull(cfg *config, w http.ResponseWriter, r *http.Request, params url.Values, parsed *listParams) {
	updated, deleted := api.sb.SubscribeType(cfg.typeName)

	prev, err := api.list(cfg, r, parsed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = writeEvent(w, "list", prev)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ticker := time.NewTicker(5 * time.Second)

	for {
		changedID := ""

		select {
		case u := <-updated:
			changedID = metadata.GetMetadata(u).ID

		case <-deleted:

		case <-ticker.C:
			err = writeEvent(w, "heartbeat", emptyEvent)
			if err != nil {
				return
			}

			continue

		case <-r.Context().Done():
			return
		}

		list, err := api.list(cfg, r, parsed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if listMatches(prev, list, changedID) {
			continue
		}

		err = writeEvent(w, "list", list)
		if err != nil {
			return
		}

		prev = list
	}
}

func (api *API) streamListDiff(cfg *config, w http.ResponseWriter, r *http.Request, params url.Values, parsed *listParams) {
	updated, deleted := api.sb.SubscribeType(cfg.typeName)

	list, err := api.list(cfg, r, parsed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	live := map[string]any{}

	for _, obj := range list {
		err = writeEvent(w, "add", obj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		live[metadata.GetMetadata(obj).ID] = obj
	}

	ticker := time.NewTicker(5 * time.Second)

	for {
		changedID := ""

		select {
		case u := <-updated:
			changedID = metadata.GetMetadata(u).ID

		case <-deleted:

		case <-ticker.C:
			err = writeEvent(w, "heartbeat", emptyEvent)
			if err != nil {
				return
			}

			continue

		case <-r.Context().Done():
			return
		}

		list, err := api.list(cfg, r, parsed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cur := map[string]any{}

		for _, obj := range list {
			objMD := metadata.GetMetadata(obj)

			_, found := live[objMD.ID]
			if found {
				// In both old and new lists, but flagged as changed by subscription chan
				if objMD.ID == changedID {
					err = writeEvent(w, "update", obj)
					if err != nil {
						return
					}
				}
			} else {
				// In the new list but not the old
				err = writeEvent(w, "add", obj)
				if err != nil {
					return
				}
			}

			cur[objMD.ID] = obj
			live[objMD.ID] = obj
		}

		for id, old := range live {
			_, found := cur[id]
			if found {
				continue
			}

			// In the old list but not the new
			err = writeEvent(w, "remove", old)
			if err != nil {
				return
			}

			delete(live, id)
		}
	}
}

func listMatches(l1 []any, l2 []any, changedID string) bool {
	if len(l1) != len(l2) {
		return false
	}

	for i, o1 := range l1 {
		o1ID := metadata.GetMetadata(o1).ID

		if changedID != "" && o1ID == changedID {
			return false
		}

		o2 := l2[i]
		o2ID := metadata.GetMetadata(o2).ID

		if o1ID != o2ID {
			return false
		}
	}

	return true
}
