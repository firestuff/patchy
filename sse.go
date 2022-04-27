package patchy

import "encoding/json"
import "fmt"
import "net/http"

var emptyEvent = map[string]string{}

func writeEvent(w http.ResponseWriter, event string, obj any) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("Failed to encode JSON: %s", err)
	}

	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	w.(http.Flusher).Flush()

	return nil
}
