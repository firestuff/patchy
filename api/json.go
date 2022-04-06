package api

import "encoding/json"
import "fmt"
import "net/http"

import "github.com/firestuff/patchy/metadata"

func readJson(r *http.Request, obj any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(obj)
}

func writeJson(w http.ResponseWriter, obj any) error {
	m := metadata.GetMetadata(obj)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, m.Sha256))

	enc := json.NewEncoder(w)
	return enc.Encode(obj)
}

func writeJsonList(w http.ResponseWriter, list []any) error {
	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	return enc.Encode(list)
}
