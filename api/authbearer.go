package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/firestuff/patchy/header"
	"github.com/firestuff/patchy/jsrest"
)

func authBearer[T any](r *http.Request, api *API, name, path string) (*http.Request, error) {
	scheme, val := header.ParseAuthorization(r)

	if strings.ToLower(scheme) != "bearer" {
		return r, nil
	}

	bearers, err := ListName[T](
		context.WithValue(r.Context(), ContextInternal, true),
		api,
		name,
		&ListOpts{
			Filters: []*Filter{
				{
					Path:  path,
					Op:    "eq",
					Value: val,
				},
			},
		},
	)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "list tokens for auth failed (%w)", err)
	}

	if len(bearers) != 1 {
		return r, nil
	}

	return r.WithContext(context.WithValue(r.Context(), ContextBearer, bearers[0])), nil
}

func SetAuthBearerName[T any](api *API, name, path string) {
	api.authBearer = func(r *http.Request, a *API) (*http.Request, error) {
		return authBearer[T](r, a, name, path)
	}
}

func SetAuthBearer[T any](api *API, path string) {
	SetAuthBearerName[T](api, objName(new(T)), path)
}
