package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/firestuff/patchy/jsrest"
)

type EmptyEventType map[string]any

var emptyEvent = EmptyEventType{}

func writeEvent(w http.ResponseWriter, event string, obj any, flush bool) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "encode JSON failed (%w)", err)
	}

	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)

	if flush {
		w.(http.Flusher).Flush()
	}

	return nil
}
