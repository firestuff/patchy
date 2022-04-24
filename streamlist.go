package api

import "net/http"
import "net/url"
import "time"

import "github.com/firestuff/patchy/metadata"

func (api *API) streamList(cfg *config, w http.ResponseWriter, r *http.Request) {
	_, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusBadRequest)
		return
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parsed, err := parseListParams(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	stream := params.Get("_stream")
	params.Del("_stream")
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

	closeChan := w.(http.CloseNotifier).CloseNotify()
	ticker := time.NewTicker(5 * time.Second)

	connected := true
	for connected {
		changedId := ""

		select {
		case u := <-updated:
			changedId = metadata.GetMetadata(u).Id

		case <-deleted:

		case <-ticker.C:
			err = writeEvent(w, "heartbeat", emptyEvent)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			continue

		case <-closeChan:
			connected = false
			continue
		}

		list, err := api.list(cfg, r, parsed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if listMatches(prev, list, changedId) {
			continue
		}

		err = writeEvent(w, "list", list)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		prev = list
	}
}

func (api *API) streamListDiff(cfg *config, w http.ResponseWriter, r *http.Request, params url.Values, parsed *listParams) {
	// XXX updated, deleted := api.sb.SubscribeType(cfg.typeName)

	list, err := api.list(cfg, r, parsed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, obj := range list {
		err = writeEvent(w, "add", obj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func listMatches(l1 []any, l2 []any, changedId string) bool {
	if len(l1) != len(l2) {
		return false
	}

	for i, o1 := range l1 {
		o1Id := metadata.GetMetadata(o1).Id

		if changedId != "" && o1Id == changedId {
			return false
		}

		o2 := l2[i]
		o2Id := metadata.GetMetadata(o2).Id

		if o1Id != o2Id {
			return false
		}
	}

	return true
}
