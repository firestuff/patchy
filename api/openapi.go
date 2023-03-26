package api

import (
	"fmt"
	"net/http"

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

func (api *API) handleOpenAPI(w http.ResponseWriter, _ *http.Request) {
	// TODO: Wrap in error writer function
	t := openapi3.T{
		OpenAPI: "3.0.3",
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{},
		},
		Paths: openapi3.Paths{},
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
			Get: &openapi3.Operation{
				Summary: fmt.Sprintf("List %s objects", name),
			},

			Post: &openapi3.Operation{
				Summary: fmt.Sprintf("Create new %s object", name),
			},
		}

		idKey := fmt.Sprintf("%s/{id}", name)
		t.Paths[idKey] = &openapi3.PathItem{
			Get: &openapi3.Operation{
				Summary: fmt.Sprintf("Get %s object", name),
			},

			Put: &openapi3.Operation{
				Summary: fmt.Sprintf("Replace %s object", name),
			},

			Patch: &openapi3.Operation{
				Summary: fmt.Sprintf("Update %s object", name),
			},

			Delete: &openapi3.Operation{
				Summary: fmt.Sprintf("Delete %s object", name),
			},
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
