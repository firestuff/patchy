package patchy

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/firestuff/patchy/metadata"
)

func readJSON(r *http.Request, obj any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	return dec.Decode(obj)
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
