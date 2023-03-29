package api

import (
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/path"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"golang.org/x/net/idna"
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
	err := api.handleOpenAPIInt(w, r)
	if err != nil {
		jsrest.WriteError(w, err)
	}
}

func (api *API) handleOpenAPIInt(w http.ResponseWriter, r *http.Request) error {
	t, err := api.buildOpenAPIGlobal(r)
	if err != nil {
		return err
	}

	names := []string{}

	for name := range api.registry {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		err = api.buildOpenAPIType(t, api.registry[name])
		if err != nil {
			return err
		}
	}

	js, err := t.MarshalJSON()
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "marshal JSON failed (%w)", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(js)

	return nil
}

func (api *API) buildOpenAPIGlobal(r *http.Request) (*openapi3.T, error) {
	baseURL, err := api.requestBaseURL(r)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "get base URL failed (%w)", err)
	}

	t := &openapi3.T{
		OpenAPI:  "3.0.3",
		Paths:    openapi3.Paths{},
		Tags:     openapi3.Tags{},
		Security: openapi3.SecurityRequirements{},

		Components: &openapi3.Components{
			RequestBodies:   openapi3.RequestBodies{},
			SecuritySchemes: openapi3.SecuritySchemes{},

			Headers: openapi3.Headers{
				"etag": &openapi3.HeaderRef{
					Value: &openapi3.Header{
						Parameter: openapi3.Parameter{
							Name: "ETag",
							In:   "header",
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: "string",
								},
							},
						},
					},
				},

				"if-match": &openapi3.HeaderRef{
					Value: &openapi3.Header{
						Parameter: openapi3.Parameter{
							Name: "If-Match",
							In:   "header",
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: "string",
								},
							},
						},
					},
				},

				"if-none-match": &openapi3.HeaderRef{
					Value: &openapi3.Header{
						Parameter: openapi3.Parameter{
							Name: "If-None-Match",
							In:   "header",
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: "string",
								},
							},
						},
					},
				},
			},

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

				"_stream": &openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "_stream",
						In:          "query",
						Description: "EventStream (List) format",
						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: "enum",
								Enum: []any{
									"full",
									"diff",
								},
							},
						},
					},
				},

				"_limit": &openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "_limit",
						In:          "query",
						Description: "Limit number of objects returned",
						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: "integer",
							},
						},
					},
				},

				"_offset": &openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "_offset",
						In:          "query",
						Description: "Skip number of objects at start of list",
						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: "integer",
							},
						},
					},
				},

				"_after": &openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "_after",
						In:          "query",
						Description: "Skip objects up to and including this ID",
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
						Type: "string",
					},
				},

				"etag": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: "string",
					},
				},

				"generation": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:   "integer",
						Format: "int64",
					},
				},

				"prefix": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: "string",
					},
				},

				"event-stream-object": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Title: "EventStream (Object)",
						Type:  "string",
						Extensions: map[string]any{
							"x-event-types": []string{
								"notModified",
								"initial",
								"update",
								"delete",
								"heartbeat",
								"error",
							},
						},
					},
				},

				"event-stream-list": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Title: "EventStream (List)",
						OneOf: openapi3.SchemaRefs{
							&openapi3.SchemaRef{
								Ref: "#/components/schemas/event-stream-list-full",
							},
							&openapi3.SchemaRef{
								Ref: "#/components/schemas/event-stream-list-diff",
							},
						},
					},
				},

				"event-stream-list-full": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Title: "EventStream (List; _stream=full)",
						Extensions: map[string]any{
							"x-event-types": []string{
								"notModified",
								"list",
								"heartbeat",
								"error",
							},
						},
					},
				},

				"event-stream-list-diff": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Title: "EventStream (List; _stream=diff)",
						Extensions: map[string]any{
							"x-event-types": []string{
								"notModified",
								"add",
								"remove",
								"update",
								"sync",
								"heartbeat",
								"error",
							},
						},
					},
				},
			},
		},

		Servers: openapi3.Servers{
			&openapi3.Server{
				URL: baseURL,
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

	if api.authBearer != nil {
		t.Components.SecuritySchemes["bearerAuth"] = &openapi3.SecuritySchemeRef{
			Value: &openapi3.SecurityScheme{
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "secret-token:*",
			},
		}

		t.Security = append(t.Security, openapi3.SecurityRequirement{"bearerAuth": []string{}})
	}

	return t, nil
}

func (api *API) buildOpenAPIType(t *openapi3.T, cfg *config) error {
	t.Tags = append(t.Tags, &openapi3.Tag{
		Name: cfg.typeName,
	})

	responseSchema, err := generateSchemaRef(cfg.typeOf)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "generate schema ref failed (%w)", err)
	}

	responseSchema.Ref = ""
	responseSchema.Value.Title = fmt.Sprintf("%s Response", cfg.typeName)

	responseSchema.Value.Properties["id"] = &openapi3.SchemaRef{Ref: "#/components/schemas/id"}
	responseSchema.Value.Properties["etag"] = &openapi3.SchemaRef{Ref: "#/components/schemas/etag"}
	responseSchema.Value.Properties["generation"] = &openapi3.SchemaRef{Ref: "#/components/schemas/generation"}

	t.Components.Schemas[fmt.Sprintf("%s--response", cfg.typeName)] = responseSchema

	requestSchema, err := generateSchemaRef(cfg.typeOf)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "generate schema ref failed (%w)", err)
	}

	requestSchema.Ref = ""
	delete(requestSchema.Value.Properties, "id")
	delete(requestSchema.Value.Properties, "etag")
	delete(requestSchema.Value.Properties, "generation")

	requestSchema.Value.Title = fmt.Sprintf("%s Request", cfg.typeName)

	t.Components.Schemas[fmt.Sprintf("%s--request", cfg.typeName)] = requestSchema

	t.Components.RequestBodies[cfg.typeName] = &openapi3.RequestBodyRef{
		Value: &openapi3.RequestBody{
			Required: true,
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{
						Ref: fmt.Sprintf("#/components/schemas/%s--request", cfg.typeName),
					},
				},
			},
		},
	}

	t.Components.Responses[cfg.typeName] = &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: P(fmt.Sprintf("OK: `%s`", cfg.typeName)),
			Headers: openapi3.Headers{
				"ETag": &openapi3.HeaderRef{
					Ref: "#/components/headers/etag",
				},
			},
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{
						Ref: fmt.Sprintf("#/components/schemas/%s--response", cfg.typeName),
					},
				},
				"text/event-stream": &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{
						Ref: "#/components/schemas/event-stream-object",
					},
				},
			},
		},
	}

	t.Components.Responses[fmt.Sprintf("%s--list", cfg.typeName)] = &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: P(fmt.Sprintf("OK: List of `%s`", cfg.typeName)),
			Headers: openapi3.Headers{
				"ETag": &openapi3.HeaderRef{
					Ref: "#/components/headers/etag",
				},
			},
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: "array",
							Items: &openapi3.SchemaRef{
								Ref: fmt.Sprintf("#/components/schemas/%s--response", cfg.typeName),
							},
						},
					},
				},
				"text/event-stream": &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{
						Ref: "#/components/schemas/event-stream-list",
					},
				},
			},
		},
	}

	paths, err := path.ListType(cfg.typeOf)
	if err != nil {
		return err
	}

	sorts := []any{}
	filters := openapi3.Parameters{}

	for _, pth := range paths {
		sorts = append(sorts, fmt.Sprintf("+%s", pth), fmt.Sprintf("-%s", pth))

		pthSchema, err := generateSchemaRef(path.GetFieldType(cfg.typeOf, pth))
		if err != nil {
			return jsrest.Errorf(jsrest.ErrInternalServerError, "generate schema ref failed (%w)", err)
		}

		filters = append(filters, openapi3.Parameters{
			&openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name:        pth,
					In:          "query",
					Description: fmt.Sprintf("Filter list by `%s` equal to", pth),
					Schema:      pthSchema,
				},
			},

			&openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name:        fmt.Sprintf("%s[gt]", pth),
					In:          "query",
					Description: fmt.Sprintf("Filter list by `%s` greater than", pth),
					Schema:      pthSchema,
				},
			},

			&openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name:        fmt.Sprintf("%s[gte]", pth),
					In:          "query",
					Description: fmt.Sprintf("Filter list by `%s` greater than or equal to", pth),
					Schema:      pthSchema,
				},
			},

			&openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name:        fmt.Sprintf("%s[hp]", pth),
					In:          "query",
					Description: fmt.Sprintf("Filter list by `%s` has prefix", pth),
					Schema: &openapi3.SchemaRef{
						Ref: "#/components/schemas/prefix",
					},
				},
			},

			&openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name:        fmt.Sprintf("%s[in]", pth),
					In:          "query",
					Description: fmt.Sprintf("Filter list by `%s` one of", pth),
					Explode:     P(false),
					Schema: &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:  "array",
							Items: pthSchema,
						},
					},
				},
			},

			&openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name:        fmt.Sprintf("%s[lt]", pth),
					In:          "query",
					Description: fmt.Sprintf("Filter list by `%s` less than", pth),
					Schema:      pthSchema,
				},
			},

			&openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					Name:        fmt.Sprintf("%s[lte]", pth),
					In:          "query",
					Description: fmt.Sprintf("Filter list by `%s` less than or equal to", pth),
					Schema:      pthSchema,
				},
			},
		}...)
	}

	t.Paths[fmt.Sprintf("/%s", cfg.typeName)] = &openapi3.PathItem{
		Get: &openapi3.Operation{
			Tags:    []string{cfg.typeName},
			Summary: fmt.Sprintf("List %s objects", cfg.typeName),
			Parameters: append(filters, openapi3.Parameters{
				&openapi3.ParameterRef{
					Ref: "#/components/headers/if-none-match",
				},
				&openapi3.ParameterRef{
					Ref: "#/components/parameters/_stream",
				},
				&openapi3.ParameterRef{
					Ref: "#/components/parameters/_limit",
				},
				&openapi3.ParameterRef{
					Ref: "#/components/parameters/_offset",
				},
				&openapi3.ParameterRef{
					Ref: "#/components/parameters/_after",
				},
				&openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "_sort",
						In:          "query",
						Description: "Direction (`+` ascending or `-` descending) and field path to sort by",
						Explode:     P(true),
						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: "array",
								Items: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: "enum",
										Enum: sorts,
									},
								},
							},
						},
					},
				},
			}...),
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{
					Ref: fmt.Sprintf("#/components/responses/%s--list", cfg.typeName),
				},
			},
		},

		Post: &openapi3.Operation{
			Tags:    []string{cfg.typeName},
			Summary: fmt.Sprintf("Create new %s object", cfg.typeName),
			RequestBody: &openapi3.RequestBodyRef{
				Ref: fmt.Sprintf("#/components/requestBodies/%s", cfg.typeName),
			},
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{
					Ref: fmt.Sprintf("#/components/responses/%s", cfg.typeName),
				},
			},
		},
	}

	t.Paths[fmt.Sprintf("/%s/{id}", cfg.typeName)] = &openapi3.PathItem{
		Parameters: openapi3.Parameters{
			&openapi3.ParameterRef{
				Ref: "#/components/parameters/id",
			},
		},

		Get: &openapi3.Operation{
			Tags:    []string{cfg.typeName},
			Summary: fmt.Sprintf("Get %s object", cfg.typeName),
			Parameters: openapi3.Parameters{
				&openapi3.ParameterRef{
					Ref: "#/components/headers/if-none-match",
				},
			},
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{
					Ref: fmt.Sprintf("#/components/responses/%s", cfg.typeName),
				},
			},
		},

		Put: &openapi3.Operation{
			Tags:    []string{cfg.typeName},
			Summary: fmt.Sprintf("Replace %s object", cfg.typeName),
			Parameters: openapi3.Parameters{
				&openapi3.ParameterRef{
					Ref: "#/components/headers/if-match",
				},
			},
			RequestBody: &openapi3.RequestBodyRef{
				Ref: fmt.Sprintf("#/components/requestBodies/%s", cfg.typeName),
			},
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{
					Ref: fmt.Sprintf("#/components/responses/%s", cfg.typeName),
				},
			},
		},

		Patch: &openapi3.Operation{
			Tags:    []string{cfg.typeName},
			Summary: fmt.Sprintf("Update %s object", cfg.typeName),
			Parameters: openapi3.Parameters{
				&openapi3.ParameterRef{
					Ref: "#/components/headers/if-match",
				},
			},
			RequestBody: &openapi3.RequestBodyRef{
				Ref: fmt.Sprintf("#/components/requestBodies/%s", cfg.typeName),
			},
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{
					Ref: fmt.Sprintf("#/components/responses/%s", cfg.typeName),
				},
			},
		},

		Delete: &openapi3.Operation{
			Tags:    []string{cfg.typeName},
			Summary: fmt.Sprintf("Delete %s object", cfg.typeName),
			Parameters: openapi3.Parameters{
				&openapi3.ParameterRef{
					Ref: "#/components/headers/if-match",
				},
			},
			Responses: openapi3.Responses{
				"204": &openapi3.ResponseRef{
					Ref: "#/components/responses/no-content",
				},
			},
		},
	}

	return nil
}

func (api *API) requestBaseURL(r *http.Request) (string, error) {
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}

	host, err := idna.ToUnicode(r.Host)
	if err != nil {
		return "", jsrest.Errorf(jsrest.ErrInternalServerError, "unicode hostname conversion failed (%w)", err)
	}

	i := strings.Index(r.RequestURI, "/_openapi")
	if i == -1 {
		return "", jsrest.Errorf(jsrest.ErrInternalServerError, "missing /_openapi in URL")
	}

	path := r.RequestURI[:i]

	return fmt.Sprintf("%s://%s%s", scheme, host, path), nil
}

func generateSchemaRef(t reflect.Type) (*openapi3.SchemaRef, error) {
	gen := openapi3gen.NewGenerator()

	schemaRef, err := gen.GenerateSchemaRef(t)
	if err != nil {
		return nil, err
	}

	for ref := range gen.SchemaRefs {
		ref.Ref = ""
	}

	return schemaRef, nil
}
