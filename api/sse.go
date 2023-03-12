package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/firestuff/patchy/jsrest"
)

var emptyEvent = map[string]string{}

func writeEvent(w http.ResponseWriter, event string, obj any) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "encode JSON failed (%w)", err)
	}

	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	w.(http.Flusher).Flush()

	return nil
}
