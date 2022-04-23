package api

import "fmt"
import "net/http"
import "os"
import "strings"

import "github.com/gorilla/mux"

import "github.com/firestuff/patchy/metadata"

func (api *API) patch(cfg *config, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj := cfg.factory()

	metadata.GetMetadata(obj).Id = vars["id"]

	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	err := api.sb.Read(cfg.typeName, obj)
	if err == os.ErrNotExist {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ifMatch := r.Header.Get("If-Match")
	if ifMatch != "" {
		if len(ifMatch) < 2 || !strings.HasPrefix(ifMatch, `"`) || !strings.HasSuffix(ifMatch, `"`) {
			http.Error(w, "Invalid If-Match", http.StatusBadRequest)
			return
		}

		if ifMatch[1:len(ifMatch)-1] != metadata.GetMetadata(obj).ETag {
			http.Error(w, fmt.Sprintf("If-Match mismatch"), http.StatusPreconditionFailed)
			return
		}
	}

	patch := cfg.factory()

	err = readJson(r, patch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Metadata is immutable or server-owned
	metadata.ClearMetadata(patch)

	if cfg.mayUpdate != nil {
		err = cfg.mayUpdate(obj, patch, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	err = merge(obj, patch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadata.GetMetadata(obj).Generation++

	err = api.sb.Write(cfg.typeName, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = writeJson(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
