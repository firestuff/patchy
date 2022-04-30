package patchy

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/firestuff/patchy/metadata"
)

func ifMatch(obj any, w http.ResponseWriter, r *http.Request) bool {
	match := r.Header.Get("If-Match")
	if match == "" {
		return true
	}

	objMD := metadata.GetMetadata(obj)

	if len(match) < 2 || !strings.HasPrefix(match, `"`) || !strings.HasSuffix(match, `"`) {
		http.Error(w, "Invalid If-Match (missing quotes)", http.StatusBadRequest)
		return false
	}

	val := strings.TrimPrefix(strings.TrimSuffix(match, `"`), `"`)

	switch {
	case strings.HasPrefix(val, "etag:"):
		if val == objMD.ETag {
			return true
		} else {
			http.Error(w, "If-Match mismatch", http.StatusPreconditionFailed)
			return false
		}

	case strings.HasPrefix(val, "generation:"):
		gen, err := strconv.ParseInt(strings.TrimPrefix(val, "generation:"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return false
		}

		if gen == objMD.Generation {
			return true
		} else {
			http.Error(w, "If-Match mismatch", http.StatusPreconditionFailed)
			return false
		}

	default:
		http.Error(w, "Invalid If-Match (unknown type)", http.StatusBadRequest)
		return false
	}
}
