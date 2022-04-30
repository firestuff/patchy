package patchy

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var emptyEvent = map[string]string{}

func writeEvent(w http.ResponseWriter, event string, obj any) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %s", err)
	}

	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	w.(http.Flusher).Flush()

	return nil
}
