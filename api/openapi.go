package api

import (
	"fmt"
	"net/http"
	"sort"

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
		Paths:   openapi3.Paths{},
		Tags:    openapi3.Tags{},

		Components: &openapi3.Components{
			Schemas:       openapi3.Schemas{},
			RequestBodies: openapi3.RequestBodies{},

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

			Responses: openapi3.Responses{
				"no-content": &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: P("No Content"),
					},
				},
			},
		},
	}

	if api.openAPI.info != nil {
		t.Info = api.openAPI.info
	}

	names := []string{}

	for name := range api.registry {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		cfg := api.registry[name]

		t.Tags = append(t.Tags, &openapi3.Tag{
			Name: name,
		})

		ref, err := openapi3gen.NewSchemaRefForValue(cfg.factory(), t.Components.Schemas)
		if err != nil {
			err = jsrest.Errorf(jsrest.ErrInternalServerError, "generate schema ref failed (%w)", err)
			jsrest.WriteError(w, err)

			return
		}

		t.Components.Schemas[name] = ref

		ref2, err := openapi3gen.NewSchemaRefForValue(cfg.factory(), t.Components.Schemas)
		if err != nil {
			err = jsrest.Errorf(jsrest.ErrInternalServerError, "generate schema ref failed (%w)", err)
			jsrest.WriteError(w, err)

			return
		}

		delete(ref2.Value.Properties, "id")
		delete(ref2.Value.Properties, "etag")
		delete(ref2.Value.Properties, "generation")

		t.Components.Schemas[fmt.Sprintf("%s--request", name)] = ref2

		t.Components.RequestBodies[name] = &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Description: name,
				Required:    true,
				Content: openapi3.Content{
					"application/json": &openapi3.MediaType{
						Schema: &openapi3.SchemaRef{
							Ref: fmt.Sprintf("#/components/schemas/%s--request", name),
						},
					},
				},
			},
		}

		t.Components.Responses[name] = &openapi3.ResponseRef{
			Value: &openapi3.Response{
				// TODO: Headers (ETag)
				Description: P("OK (Object)"),
				Content: openapi3.Content{
					"application/json": &openapi3.MediaType{
						Schema: &openapi3.SchemaRef{
							Ref: fmt.Sprintf("#/components/schemas/%s", name),
						},
					},
				},
			},
		}

		t.Components.Responses[fmt.Sprintf("%s--list", name)] = &openapi3.ResponseRef{
			Value: &openapi3.Response{
				// TODO: Headers (ETag)
				Description: P("OK (Array)"),
				Content: openapi3.Content{
					"application/json": &openapi3.MediaType{
						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Description: fmt.Sprintf("Array of %s", name),
								Type:        "array",
								Items: &openapi3.SchemaRef{
									Ref: fmt.Sprintf("#/components/schemas/%s", name),
								},
							},
						},
					},
				},
			},
		}

		t.Paths[fmt.Sprintf("/%s", name)] = &openapi3.PathItem{
			Get: &openapi3.Operation{
				Tags:    []string{name, "list"},
				Summary: fmt.Sprintf("List %s objects", name),
				Responses: openapi3.Responses{
					"200": &openapi3.ResponseRef{
						Ref: fmt.Sprintf("#/components/responses/%s--list", name),
					},
				},
			},

			Post: &openapi3.Operation{
				Tags:    []string{name, "create"},
				Summary: fmt.Sprintf("Create new %s object", name),
				RequestBody: &openapi3.RequestBodyRef{
					Ref: fmt.Sprintf("#/components/requestBodies/%s", name),
				},
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
				Tags:    []string{name, "get"},
				Summary: fmt.Sprintf("Get %s object", name),
				Responses: openapi3.Responses{
					"200": &openapi3.ResponseRef{
						Ref: fmt.Sprintf("#/components/responses/%s", name),
					},
				},
			},

			Put: &openapi3.Operation{
				Tags:    []string{name, "replace"},
				Summary: fmt.Sprintf("Replace %s object", name),
				RequestBody: &openapi3.RequestBodyRef{
					Ref: fmt.Sprintf("#/components/requestBodies/%s", name),
				},
				Responses: openapi3.Responses{
					"200": &openapi3.ResponseRef{
						Ref: fmt.Sprintf("#/components/responses/%s", name),
					},
				},
			},

			Patch: &openapi3.Operation{
				Tags:    []string{name, "update"},
				Summary: fmt.Sprintf("Update %s object", name),
				RequestBody: &openapi3.RequestBodyRef{
					Ref: fmt.Sprintf("#/components/requestBodies/%s", name),
				},
				Responses: openapi3.Responses{
					"200": &openapi3.ResponseRef{
						Ref: fmt.Sprintf("#/components/responses/%s", name),
					},
				},
			},

			Delete: &openapi3.Operation{
				Tags:    []string{name, "delete"},
				Summary: fmt.Sprintf("Delete %s object", name),
				Responses: openapi3.Responses{
					"204": &openapi3.ResponseRef{
						Ref: "#/components/responses/no-content",
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
