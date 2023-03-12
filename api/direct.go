package api

import (
	"context"
	"fmt"

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

func FindName[T any](ctx context.Context, api *API, name, shortID string) (*T, error) {
	listOpts := &ListOpts{
		Filters: []*Filter{
			{
				Path:  "id",
				Op:    "hp",
				Value: shortID,
			},
		},
	}

	objs, err := ListName[T](ctx, api, name, listOpts)
	if err != nil {
		return nil, err
	}

	if len(objs) == 0 {
		return nil, fmt.Errorf("no object found with short ID: %s", shortID)
	}

	if len(objs) > 1 {
		return nil, fmt.Errorf("multiple objects found with short ID: %s", shortID)
	}

	return objs[0], nil
}

func Find[T any](ctx context.Context, api *API, shortID string) (*T, error) {
	return FindName[T](ctx, api, objName(new(T)), shortID)
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

func UpdateName[T any](ctx context.Context, api *API, name, id string, obj *T) (*T, error) {
	cfg := api.registry[name]
	if cfg == nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "unknown type: %s", name)
	}

	updated, err := api.updateInt(ctx, cfg, nil, id, obj)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "update failed (%w)", err)
	}

	return updated.(*T), nil
}

func Update[T any](ctx context.Context, api *API, id string, obj *T) (*T, error) {
	return UpdateName[T](ctx, api, objName(obj), id, obj)
}
