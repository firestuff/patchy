package patchy

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/firestuff/patchy/jsrest"
)

var emptyEvent = map[string]string{}

func writeEvent(w http.ResponseWriter, event string, obj any) *jsrest.Error {
	data, err := json.Marshal(obj)
	if err != nil {
		e := fmt.Errorf("failed to encode JSON: %w", err)
		jse := jsrest.FromError(e, jsrest.StatusInternalServerError)

		return jse
	}

	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	w.(http.Flusher).Flush()

	return nil
}
