package api

import (
	"context"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
	"github.com/google/uuid"
)

func CreateName[T any](ctx context.Context, api *API, name string, obj *T) (*T, error) {
	cfg := api.registry[name]
	if cfg == nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "unknown type: %s", name)
	}

	metadata.GetMetadata(obj).ID = uuid.NewString()

	err := api.sb.Write(ctx, cfg.typeName, obj)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "write failed (%w)", err)
	}

	return obj, nil
}

func Create[T any](ctx context.Context, api *API, obj *T) (*T, error) {
	return CreateName[T](ctx, api, objName(obj), obj)
}

func GetName[T any](ctx context.Context, api *API, name, id string) (*T, error) {
	cfg := api.registry[name]
	if cfg == nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "unknown type: %s", name)
	}

	obj, err := api.sb.Read(ctx, cfg.typeName, id, cfg.factory)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	return obj.(*T), nil
}

func Get[T any](ctx context.Context, api *API, id string) (*T, error) {
	return GetName[T](ctx, api, objName(new(T)), id)
}

/*
func (api *API) List(name string, params *ListOpts) ([]any, error) {
	// TODO: Expose listParams as ListOps and filter as Match, like client
	// TODO: Probably import those into client
}
*/
