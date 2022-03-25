package api

import "encoding/json"
import "net/http"

func readJson(r *http.Request, obj any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(obj)
}

func writeJson(w http.ResponseWriter, obj any) error {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	return enc.Encode(obj)
}
