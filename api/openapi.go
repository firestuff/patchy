package api

import (
	"net/http"

	"github.com/firestuff/patchy/jsrest"
	"github.com/getkin/kin-openapi/openapi3"
)

func (api *API) handleOpenAPI(w http.ResponseWriter, _ *http.Request) {
	t := openapi3.T{
		OpenAPI: "3.0.3",
	}

	err := jsrest.Write(w, t)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "write failed (%w)", err)
		jsrest.WriteError(w, err)
	}
}
