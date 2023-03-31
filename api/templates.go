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

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/path"
	"github.com/julienschmidt/httprouter"
)

//go:embed templates/*
var templateFS embed.FS

var templates = template.Must(
	template.New("templates").
		Funcs(template.FuncMap{
			"padRight": padRight,
		}).
		ParseFS(templateFS, "templates/*"))

type templateInput struct {
	Form  url.Values
	Types []*templateType
}

type templateType struct {
	APIName      string
	GoName       string
	Fields       []*templateField
	GoNameMaxLen int
	GoTypeMaxLen int
}

type templateField struct {
	APIName string
	GoName  string
	GoType  string
}

func (api *API) registerTemplates() {
	api.router.GET("/_goclient", api.writeTemplate("goclient"))
	api.router.GET("/_tsclient", api.writeTemplate("tsclient"))
}

func (api *API) writeTemplate(name string) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		err := r.ParseForm()
		if err != nil {
			err = jsrest.Errorf(jsrest.ErrBadRequest, "parse params failed (%w)", err)
			jsrest.WriteError(w, err)

			return
		}

		input := &templateInput{
			Form: r.Form,
		}

		for _, name := range api.names() {
			cfg := api.registry[name]

			tt := &templateType{
				APIName: name,
				GoName:  upperFirst(cfg.typeOf.Name()),
			}

			path.WalkType(cfg.typeOf, func(_ string, parts []string, field reflect.StructField) {
				// TODO: Support nested structs
				tf := &templateField{
					APIName: parts[0],
					GoName:  upperFirst(field.Name),
					GoType:  field.Type.String(),
				}

				if len(tf.GoName) > tt.GoNameMaxLen {
					tt.GoNameMaxLen = len(tf.GoName)
				}

				if len(tf.GoType) > tt.GoTypeMaxLen {
					tt.GoTypeMaxLen = len(tf.GoType)
				}

				tt.Fields = append(tt.Fields, tf)
			})

			input.Types = append(input.Types, tt)
		}

		// Buffer this so we can handle the error before sending any template output
		buf := &bytes.Buffer{}

		err = templates.ExecuteTemplate(buf, name, input)
		if err != nil {
			err = jsrest.Errorf(jsrest.ErrInternalServerError, "execute template failed (%w)", err)
			jsrest.WriteError(w, err)

			return
		}

		w.Header().Set("Content-Type", "text/plain")

		_, _ = buf.WriteTo(w)
	}
}

func padRight(in string, l int) string {
	return fmt.Sprintf(fmt.Sprintf("%%-%ds", l), in)
}

func upperFirst(in string) string {
	return strings.ToUpper(in[:1]) + in[1:]
}
