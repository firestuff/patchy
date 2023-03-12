package api

import (
	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
	"github.com/google/uuid"
)

// TODO: Add generic wrappers with and without names

func (api *API) Create(name string, obj any) (any, error) {
	cfg := api.registry[name]
	if cfg == nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "unknown type: %s", name)
	}

	metadata.GetMetadata(obj).ID = uuid.NewString()

	err := api.sb.Write(cfg.typeName, obj)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "write failed (%w)", err)
	}

	return obj, nil
}

func (api *API) Get(name, id string) (any, error) {
	// TODO: Take ctx, pass it all the way down to db.ExecContext()
	cfg := api.registry[name]
	if cfg == nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "unknown type: %s", name)
	}

	obj, err := api.sb.Read(cfg.typeName, id, cfg.factory)
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
