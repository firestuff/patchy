package api

import (
	"context"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
	"github.com/google/uuid"
)

// TODO: Add generic wrappers with and without names

func (api *API) Create(ctx context.Context, name string, obj any) (any, error) {
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

func (api *API) Get(ctx context.Context, name, id string) (any, error) {
	cfg := api.registry[name]
	if cfg == nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "unknown type: %s", name)
	}

	obj, err := api.sb.Read(ctx, cfg.typeName, id, cfg.factory)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	return obj, nil
}

/*
func (api *API) List(name string, params *ListOpts) ([]any, error) {
	// TODO: Expose listParams as ListOps and filter as Match, like client
	// TODO: Probably import those into client
}
*/
