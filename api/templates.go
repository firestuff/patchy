package api

import (
	"bytes"
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"text/template"
	"time"

	"cloud.google.com/go/civil"
	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/path"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/exp/slices"
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
	Form       url.Values
	Types      []*templateType
	Packages   []string
	URLPrefix  string
	AuthBasic  bool
	AuthBearer bool
}

type templateType struct {
	APIName      string
	GoName       string
	Fields       []*templateField
	GoNameMaxLen int
	GoTypeMaxLen int

	typeOf reflect.Type
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
			Form:       r.Form,
			URLPrefix:  api.prefix,
			AuthBasic:  api.authBasic != nil,
			AuthBearer: api.authBearer != nil,
		}

		typeQueue := []*templateType{}
		typesDone := map[reflect.Type]bool{}

		for _, name := range api.names() {
			cfg := api.registry[name]

			typeQueue = append(typeQueue, &templateType{
				APIName: name,
				typeOf:  cfg.typeOf,
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

			path.WalkType(tt.typeOf, func(_ string, parts []string, field reflect.StructField) {
				typeOf := path.MaybeIndirectType(field.Type)

				if len(parts) > 1 {
					return
				}

				tf := &templateField{
					APIName: parts[0],
					GoName:  upperFirst(field.Name),
					GoType:  goType(field.Type),
				}

				if typeOf.Kind() == reflect.Struct && typeOf != reflect.TypeOf(time.Time{}) && typeOf != reflect.TypeOf(civil.Date{}) {
					typeQueue = append(typeQueue, &templateType{
						typeOf: typeOf,
					})
				}

				if len(tf.GoName) > tt.GoNameMaxLen {
					tt.GoNameMaxLen = len(tf.GoName)
				}

				if len(tf.GoType) > tt.GoTypeMaxLen {
					tt.GoTypeMaxLen = len(tf.GoType)
				}

				switch typeOf {
				case reflect.TypeOf(time.Time{}):
					input.addPackage("time")

				case reflect.TypeOf(civil.Date{}):
					input.addPackage("cloud.google.com/go/civil")
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

func (ti *templateInput) addPackage(name string) {
	if slices.Contains(ti.Packages, name) {
		return
	}

	ti.Packages = append(ti.Packages, name)
	sort.Strings(ti.Packages)
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
