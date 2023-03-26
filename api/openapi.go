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
			Parameters: openapi3.ParametersMap{
				"id": &openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "id",
						In:          "path",
						Description: "Object ID",
						Required:    true,
						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: "string",
							},
						},
					},
				},
			},
			Schemas:       openapi3.Schemas{},
			RequestBodies: openapi3.RequestBodies{},
			Responses:     openapi3.Responses{},
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

		t.Components.Responses[name] = &openapi3.ResponseRef{
			Value: &openapi3.Response{
				// TODO: Headers (ETag)
				Description: P("OK"),
				Content: openapi3.Content{
					"application/json": &openapi3.MediaType{
						Schema: &openapi3.SchemaRef{
							Ref: fmt.Sprintf("#/components/schemas/%s", name),
						},
					},
				},
			},
		}

		t.Paths[fmt.Sprintf("/%s", name)] = &openapi3.PathItem{
			Get: &openapi3.Operation{
				Summary: fmt.Sprintf("List %s objects", name),
				Responses: openapi3.Responses{
					"200": &openapi3.ResponseRef{
						// TODO: Make this a list of objects, not a single
						Ref: fmt.Sprintf("#/components/responses/%s", name),
					},
				},
			},

			Post: &openapi3.Operation{
				Summary: fmt.Sprintf("Create new %s object", name),
				Responses: openapi3.Responses{
					"200": &openapi3.ResponseRef{
						Ref: fmt.Sprintf("#/components/responses/%s", name),
					},
				},
			},
		}

		t.Paths[fmt.Sprintf("/%s/{id}", name)] = &openapi3.PathItem{
			Parameters: openapi3.Parameters{
				&openapi3.ParameterRef{
					Ref: "#/components/parameters/id",
				},
			},

			Get: &openapi3.Operation{
				Summary: fmt.Sprintf("Get %s object", name),
				Responses: openapi3.Responses{
					"200": &openapi3.ResponseRef{
						Ref: fmt.Sprintf("#/components/responses/%s", name),
					},
				},
			},

			Put: &openapi3.Operation{
				Summary: fmt.Sprintf("Replace %s object", name),
				Responses: openapi3.Responses{
					"200": &openapi3.ResponseRef{
						Ref: fmt.Sprintf("#/components/responses/%s", name),
					},
				},
			},

			Patch: &openapi3.Operation{
				Summary: fmt.Sprintf("Update %s object", name),
				Responses: openapi3.Responses{
					"200": &openapi3.ResponseRef{
						Ref: fmt.Sprintf("#/components/responses/%s", name),
					},
				},
			},

			Delete: &openapi3.Operation{
				Summary: fmt.Sprintf("Delete %s object", name),
				Responses: openapi3.Responses{
					"200": &openapi3.ResponseRef{
						// TODO: Make this an empty response
						Ref: fmt.Sprintf("#/components/responses/%s", name),
					},
				},
			},
		}
	}

	/*
		err = t.Validate(r.Context())
		if err != nil {
			err = jsrest.Errorf(jsrest.ErrInternalServerError, "validation failed (%w)", err)
			jsrest.WriteError(w, err)

			return
		}
	*/

	w.Header().Set("Content-Type", "application/json")

	js, err := t.MarshalJSON()
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "marshal JSON failed (%w)", err)
		jsrest.WriteError(w, err)

		return
	}

	_, _ = w.Write(js)
}
