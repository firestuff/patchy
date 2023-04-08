package api

import (
	"bytes"
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"text/template"
	"time"

	"cloud.google.com/go/civil"
	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/path"
	"github.com/julienschmidt/httprouter"
)

//go:embed templates/*
var templateFS embed.FS

var templates = template.Must(
	template.New("templates").
		Funcs(template.FuncMap{
			"add":        add,
			"padRight":   padRight,
			"upperFirst": upperFirst,
		}).
		ParseFS(templateFS, "templates/*"))

type templateInput struct {
	Info       *OpenAPIInfo
	Form       url.Values
	Types      []*templateType
	UsesTime   bool
	UsesCivil  bool
	URLPrefix  string
	AuthBasic  bool
	AuthBearer bool
}

type templateType struct {
	APIName      string
	APINameCamel string
	GoName       string
	Fields       []*templateField
	GoNameMaxLen int
	GoTypeMaxLen int
	SkipInGo     bool

	typeOf      reflect.Type
	allOptional bool
}

type templateField struct {
	APIName  string
	GoName   string
	GoType   string
	TSType   string
	Optional bool
}

func (api *API) registerTemplates() {
	api.router.GET("/_client.go", api.writeTemplate("client.go.tmpl"))
	api.router.GET("/_client.ts", api.writeTemplate("client.ts.tmpl"))
}

func (api *API) writeTemplate(name string) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		input := &templateInput{
			Info:       api.openAPI.info,
			Form:       r.Form,
			URLPrefix:  api.prefix,
			AuthBasic:  api.authBasic != nil,
			AuthBearer: api.authBearer != nil,
		}

		typeQueue := []*templateType{
			{
				typeOf:   reflect.TypeOf(DebugInfo{}),
				SkipInGo: true,
			},
			{
				typeOf:   reflect.TypeOf(jsrest.JSONError{}),
				SkipInGo: true,
			},
			{
				typeOf:      reflect.TypeOf(GetOpts{}),
				SkipInGo:    true,
				allOptional: true,
			},
			{
				typeOf:      reflect.TypeOf(ListOpts{}),
				SkipInGo:    true,
				allOptional: true,
			},
			{
				typeOf:      reflect.TypeOf(UpdateOpts{}),
				SkipInGo:    true,
				allOptional: true,
			},
		}
		typesDone := map[reflect.Type]bool{}

		for _, name := range api.names() {
			cfg := api.registry[name]

			typeQueue = append(typeQueue, &templateType{
				APIName:     name,
				typeOf:      cfg.typeOf,
				allOptional: true,
			})
		}

		for len(typeQueue) > 0 {
			tt := typeQueue[0]
			typeQueue = typeQueue[1:]

			if typesDone[tt.typeOf] {
				continue
			}

			typesDone[tt.typeOf] = true

			tt.GoName = upperFirst(tt.typeOf.Name())

			if tt.APIName != "" {
				tt.APINameCamel = tt.APIName

				if strings.EqualFold(tt.APINameCamel, tt.GoName) {
					// Pick up real camel case if we have it
					tt.APINameCamel = tt.GoName
				}

				// Standardize on first upper
				tt.APINameCamel = upperFirst(tt.APINameCamel)
			}

			path.WalkType(tt.typeOf, func(_ string, parts []string, field reflect.StructField) {
				typeOf := path.MaybeIndirectType(field.Type)

				elemType := typeOf
				if elemType.Kind() == reflect.Slice {
					elemType = path.MaybeIndirectType(elemType.Elem())
				}

				if len(parts) > 1 || parts[0] == "" {
					return
				}

				tf := &templateField{
					APIName:  parts[0],
					GoName:   upperFirst(field.Name),
					GoType:   goType(field.Type),
					TSType:   tsType(field.Type),
					Optional: tt.allOptional || field.Type.Kind() == reflect.Pointer,
				}

				if elemType.Kind() == reflect.Struct && elemType != reflect.TypeOf(time.Time{}) && elemType != reflect.TypeOf(civil.Date{}) {
					typeQueue = append(typeQueue, &templateType{
						typeOf:   elemType,
						SkipInGo: tt.SkipInGo,
					})
				}

				if len(tf.GoName) > tt.GoNameMaxLen {
					tt.GoNameMaxLen = len(tf.GoName)
				}

				if len(tf.GoType) > tt.GoTypeMaxLen {
					tt.GoTypeMaxLen = len(tf.GoType)
				}

				switch typeOf {
				case path.TimeTimeType:
					input.UsesTime = true

				case path.CivilDateType:
					input.UsesCivil = true
				}

				tt.Fields = append(tt.Fields, tf)
			})

			input.Types = append(input.Types, tt)
		}

		// Buffer this so we can handle the error before sending any template output
		buf := &bytes.Buffer{}

		err := templates.ExecuteTemplate(buf, name, input)
		if err != nil {
			err = jsrest.Errorf(jsrest.ErrInternalServerError, "execute template failed (%w)", err)
			jsrest.WriteError(w, err)

			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		_, _ = buf.WriteTo(w)
	}
}

func add(a, b int) int {
	return a + b
}

func padRight(in string, l int) string {
	return fmt.Sprintf(fmt.Sprintf("%%-%ds", l), in)
}

func upperFirst(in string) string {
	return strings.ToUpper(in[:1]) + in[1:]
}

func goType(t reflect.Type) string {
	elemType := path.MaybeIndirectType(t)

	if elemType.Kind() != reflect.Struct || elemType == path.TimeTimeType || elemType == path.CivilDateType {
		return t.String()
	}

	if t.Kind() == reflect.Pointer {
		return fmt.Sprintf("*%s", upperFirst(elemType.Name()))
	}

	return upperFirst(elemType.Name())
}

func tsType(t reflect.Type) string {
	elemType := path.MaybeIndirectType(t)

	if elemType == path.TimeTimeType || elemType == path.CivilDateType {
		return "string"
	}

	// TODO: Handle http.Header (map[string][]string) for DebugInfo

	switch elemType.Kind() { //nolint:exhaustive
	case reflect.Slice:
		return fmt.Sprintf("%s[]", tsType(elemType.Elem()))

	case reflect.Int:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		return "number"

	case reflect.Bool:
		return "boolean"

	case reflect.Struct:
		return goType(elemType)

	case reflect.Interface:
		return "any"

	default:
		return "string"
	}
}
