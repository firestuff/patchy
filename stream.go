package patchy

import "net/http"
import "os"
import "time"

import "github.com/gorilla/mux"

import "github.com/firestuff/patchy/metadata"

func (api *API) stream(cfg *config, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	_, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	obj := cfg.factory()

	metadata.GetMetadata(obj).Id = vars["id"]

	cfg.mu.RLock()
	// THIS LOCK REQUIRES MANUAL UNLOCKING IN ALL BRANCHES

	err := api.sb.Read(cfg.typeName, obj)
	if err == os.ErrNotExist {
		http.Error(w, err.Error(), http.StatusNotFound)
		cfg.mu.RUnlock()
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		cfg.mu.RUnlock()
		return
	}

	if cfg.mayRead != nil {
		err = cfg.mayRead(obj, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			cfg.mu.RUnlock()
			return
		}
	}

	err = writeEvent(w, "initial", obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		cfg.mu.RUnlock()
		return
	}

	objChan := api.sb.SubscribeKey(cfg.typeName, vars["id"])
	ticker := time.NewTicker(5 * time.Second)

	cfg.mu.RUnlock()

	for {
		select {

		case <-r.Context().Done():
			return

		case msg, ok := <-objChan:
			if ok {
				err = writeEvent(w, "update", msg)
				if err != nil {
					return
				}
			} else {
				writeEvent(w, "delete", emptyEvent)
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
