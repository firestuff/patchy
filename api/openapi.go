package api

import (
	"fmt"
	"net/http"
	"sort"
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
	prefix := strings.TrimSuffix(r.RequestURI, "/_openapi")

	t := openapi3.T{
		OpenAPI:  "3.0.3",
		Paths:    openapi3.Paths{},
		Tags:     openapi3.Tags{},
		Security: openapi3.SecurityRequirements{},

		Components: &openapi3.Components{
			RequestBodies:   openapi3.RequestBodies{},
			SecuritySchemes: openapi3.SecuritySchemes{},

			Parameters: openapi3.ParametersMap{
				"id": &openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "id",
						In:          "path",
						Description: "Object ID",
						Required:    true,
						Schema: &openapi3.SchemaRef{
							Ref: "#/components/schemas/id",
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

			Schemas: openapi3.Schemas{
				"id": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:    "string",
						Example: "SK7rZ3j13PJpeIiS",
					},
				},

				"etag": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:    "string",
						Example: "etag:20af52b66d85b8854183c82e462771a01606b31104a44a52237e17f6548d4ba7",
					},
				},

				"generation": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:    "integer",
						Format:  "int64",
						Example: 42,
					},
				},
			},
		},
	}

	if api.openAPI.info != nil {
		t.Info = api.openAPI.info
	}

	if api.authBasic {
		t.Components.SecuritySchemes["basicAuth"] = &openapi3.SecuritySchemeRef{
			Value: &openapi3.SecurityScheme{
				Type:   "http",
				Scheme: "basic",
			},
		}

		t.Security = append(t.Security, openapi3.SecurityRequirement{"basicAuth": []string{}})
	}

	if api.authBearer {
		t.Components.SecuritySchemes["bearerAuth"] = &openapi3.SecuritySchemeRef{
			Value: &openapi3.SecurityScheme{
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "secret-token:*",
			},
		}

		t.Security = append(t.Security, openapi3.SecurityRequirement{"bearerAuth": []string{}})
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

		ref.Value.Title = fmt.Sprintf("%s Response", name)

		ref.Value.Properties["id"] = &openapi3.SchemaRef{Ref: "#/components/schemas/id"}
		ref.Value.Properties["etag"] = &openapi3.SchemaRef{Ref: "#/components/schemas/etag"}
		ref.Value.Properties["generation"] = &openapi3.SchemaRef{Ref: "#/components/schemas/generation"}

		t.Components.Schemas[fmt.Sprintf("%s--response", name)] = ref

		ref2, err := openapi3gen.NewSchemaRefForValue(cfg.factory(), t.Components.Schemas)
		if err != nil {
			err = jsrest.Errorf(jsrest.ErrInternalServerError, "generate schema ref failed (%w)", err)
			jsrest.WriteError(w, err)

			return
		}

		delete(ref2.Value.Properties, "id")
		delete(ref2.Value.Properties, "etag")
		delete(ref2.Value.Properties, "generation")

		ref2.Value.Title = fmt.Sprintf("%s Request", name)

		t.Components.Schemas[fmt.Sprintf("%s--request", name)] = ref2

		t.Components.RequestBodies[name] = &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Required: true,
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
							Ref: fmt.Sprintf("#/components/schemas/%s--response", name),
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
								Type: "array",
								Items: &openapi3.SchemaRef{
									Ref: fmt.Sprintf("#/components/schemas/%s--response", name),
								},
							},
						},
					},
				},
			},
		}

		t.Paths[fmt.Sprintf("%s/%s", prefix, name)] = &openapi3.PathItem{
			Get: &openapi3.Operation{
				Tags:    []string{name, "List"},
				Summary: fmt.Sprintf("List %s objects", name),
				Responses: openapi3.Responses{
					"200": &openapi3.ResponseRef{
						Ref: fmt.Sprintf("#/components/responses/%s--list", name),
					},
				},
			},

			Post: &openapi3.Operation{
				Tags:    []string{name, "Create"},
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

		t.Paths[fmt.Sprintf("%s/%s/{id}", prefix, name)] = &openapi3.PathItem{
			Parameters: openapi3.Parameters{
				&openapi3.ParameterRef{
					Ref: "#/components/parameters/id",
				},
			},

			Get: &openapi3.Operation{
				Tags:    []string{name, "Get"},
				Summary: fmt.Sprintf("Get %s object", name),
				Responses: openapi3.Responses{
					"200": &openapi3.ResponseRef{
						Ref: fmt.Sprintf("#/components/responses/%s", name),
					},
				},
			},

			Put: &openapi3.Operation{
				Tags:    []string{name, "Replace"},
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
				Tags:    []string{name, "Update"},
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
				Tags:    []string{name, "Delete"},
				Summary: fmt.Sprintf("Delete %s object", name),
				Responses: openapi3.Responses{
					"204": &openapi3.ResponseRef{
						Ref: "#/components/responses/no-content",
					},
				},
			},
		}
	}

	js, err := t.MarshalJSON()
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "marshal JSON failed (%w)", err)
		jsrest.WriteError(w, err)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(js)
}
