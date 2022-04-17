package api

import "net/http"
import "time"

import "github.com/firestuff/patchy/metadata"

func (api *API) streamList(cfg *config, w http.ResponseWriter, r *http.Request) {
	_, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	updated, deleted := api.sb.SubscribeType(cfg.typeName)

	prev, err := api.list(cfg, r)
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
		var changed *any

		select {
		case <-closeChan:
			connected = false

		case u := <-updated:
			changed = &u

		case <-deleted:

		case <-ticker.C:
			err = writeEvent(w, "heartbeat", emptyEvent)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			continue
		}

		list, err := api.list(cfg, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if listMatches(prev, list, changed) {
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

func listMatches(l1 []any, l2 []any, changed *any) bool {
	if len(l1) != len(l2) {
		return false
	}

	var changedId string

	if changed != nil {
		changedId = metadata.GetMetadata(changed).Id
	}

	for i, o1 := range l1 {
		o1Id := metadata.GetMetadata(o1).Id

		if changed != nil && o1Id == changedId {
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