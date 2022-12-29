package patchy

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/firestuff/patchy/metadata"
)

func readJSON(w http.ResponseWriter, r *http.Request, obj any) bool {
	// TODO: Parse semicolon params
	switch r.Header.Get("Content-Type") {
	case "":
		fallthrough
	case "application/json":
		break

	default:
		http.Error(w, "unknown Content-Type", http.StatusUnsupportedMediaType)
		return false
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

func writeJSON(w http.ResponseWriter, obj any) error {
	m := metadata.GetMetadata(obj)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, m.ETag))

	enc := json.NewEncoder(w)

	return enc.Encode(obj)
}

func writeJSONList(w http.ResponseWriter, list []any) error {
	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)

	return enc.Encode(list)
}
