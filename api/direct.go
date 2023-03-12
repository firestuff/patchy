package api

import (
	"context"

	"github.com/firestuff/patchy/jsrest"
)

func CreateName[T any](ctx context.Context, api *API, name string, obj *T) (*T, error) {
	cfg := api.registry[name]
	if cfg == nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "unknown type: %s", name)
	}

	created, err := api.createInt(ctx, cfg, nil, obj)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "create failed (%w)", err)
	}

	return created.(*T), nil
}

func Create[T any](ctx context.Context, api *API, obj *T) (*T, error) {
	return CreateName[T](ctx, api, objName(obj), obj)
}

func GetName[T any](ctx context.Context, api *API, name, id string) (*T, error) {
	cfg := api.registry[name]
	if cfg == nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "unknown type: %s", name)
	}

	obj, err := api.getInt(ctx, cfg, nil, id)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "get failed (%w)", err)
	}

	return obj.(*T), nil
}

func Get[T any](ctx context.Context, api *API, id string) (*T, error) {
	return GetName[T](ctx, api, objName(new(T)), id)
}

func ListName[T any](ctx context.Context, api *API, name string, opts *ListOpts) ([]*T, error) {
	cfg := api.registry[name]
	if cfg == nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "unknown type: %s", name)
	}

	list, err := api.listInt(ctx, cfg, nil, opts)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "list failed (%w)", err)
	}

	ret := []*T{}
	for _, obj := range list {
		ret = append(ret, obj.(*T))
	}

	return ret, nil
}

func List[T any](ctx context.Context, api *API, opts *ListOpts) ([]*T, error) {
	return ListName[T](ctx, api, objName(new(T)), opts)
}
