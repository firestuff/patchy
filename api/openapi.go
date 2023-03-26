package api

import (
	"net/http"
	"strings"

	"github.com/firestuff/patchy/jsrest"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
)

type (
	OpenAPI     = openapi3.T
	OpenAPIInfo = openapi3.Info
)

type openAPI struct {
	info *OpenAPIInfo
}

func (api *API) SetOpenAPIInfo(info *OpenAPIInfo) {
	api.openAPI.info = info
}

func (api *API) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	// TODO: Wrap in error writer function
	comp := openapi3.NewComponents()
	comp.Schemas = openapi3.Schemas{}

	t := openapi3.T{
		OpenAPI:    "3.0.3",
		Components: &comp,
		Paths:      openapi3.Paths{},
		Servers: openapi3.Servers{
			&openapi3.Server{
				URL: strings.TrimSuffix(r.URL.Path, "_openapi"),
			},
		},
	}

	if api.openAPI.info != nil {
		t.Info = api.openAPI.info
	}

	for name, cfg := range api.registry {
		ref, err := openapi3gen.NewSchemaRefForValue(cfg.factory(), t.Components.Schemas)
		if err != nil {
			err = jsrest.Errorf(jsrest.ErrInternalServerError, "write failed (%w)", err)
			jsrest.WriteError(w, err)

			return
		}

		t.Components.Schemas[name] = ref

		t.Paths[name] = &openapi3.PathItem{
			Get: &openapi3.Operation{},
		}
	}

	w.Header().Set("Content-Type", "application/json")

	js, err := t.MarshalJSON()
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "marshal JSON failed (%w)", err)
		jsrest.WriteError(w, err)

		return
	}

	_, _ = w.Write(js)
}
