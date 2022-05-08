package patchy

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func (api *API) stream(cfg *config, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if _, ok := w.(http.Flusher); !ok {
		http.Error(w, "Streaming not supported", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	v, err := api.sb.Read(r.Context(), cfg.typeName, vars["id"], cfg.factory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	obj := <-v.Chan()
	if obj == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if cfg.mayRead != nil {
		err = cfg.mayRead(obj, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	err = writeEvent(w, "initial", obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-r.Context().Done():
			return

		case msg, ok := <-v.Chan():
			if ok {
				err = writeEvent(w, "update", msg)
				if err != nil {
					return
				}
			} else {
				_ = writeEvent(w, "delete", emptyEvent)
				return
			}

		case <-ticker.C:
			err = writeEvent(w, "heartbeat", emptyEvent)
			if err != nil {
				return
			}
		}
	}
}
