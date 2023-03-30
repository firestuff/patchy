package api

import (
	"bytes"
	"embed"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/firestuff/patchy/jsrest"
	"github.com/julienschmidt/httprouter"
)

//go:embed templates/*
var templateFS embed.FS
var templates = template.Must(template.ParseFS(templateFS, "templates/*"))

type templateInput struct {
	Form  url.Values
	Types []*templateType
}

type templateType struct {
	EndpointName string
	GoName       string
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

			input.Types = append(input.Types, &templateType{
				EndpointName: name,
				GoName:       upperFirst(cfg.typeOf.Name()),
			})
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

func upperFirst(in string) string {
	return strings.ToUpper(in[:1]) + in[1:]
}
