package api

import "encoding/json"
import "fmt"
import "net/http"

func writeUpdate(w http.ResponseWriter, obj any) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("Failed to encode JSON: %s", err)
	}

	fmt.Fprintf(w, "event: update\ndata: %s\n\n", data)
	w.(http.Flusher).Flush()

	return nil
}

func writeDelete(w http.ResponseWriter) {
	fmt.Fprintf(w, "event: delete\ndata: {}\n\n")
	w.(http.Flusher).Flush()
}

func writeHeartbeat(w http.ResponseWriter) {
	fmt.Fprintf(w, "event: heartbeat\ndata: {}\n\n")
	w.(http.Flusher).Flush()
}
