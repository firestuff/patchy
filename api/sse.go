package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/firestuff/patchy/jsrest"
)

type EmptyEventType map[string]any

var emptyEvent = EmptyEventType{}

func writeEvent(w http.ResponseWriter, event, id string, obj any, flush bool) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "encode JSON failed (%w)", err)
	}

	if id == "" {
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	} else {
		fmt.Fprintf(w, "event: %s\nid: %s\ndata: %s\n\n", event, id, data)
	}

	if flush {
		w.(http.Flusher).Flush()
	}

	return nil
}
