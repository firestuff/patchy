package api

import (
	"context"
	"net/http"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
	"github.com/google/uuid"
)

func (api *API) createInt(ctx context.Context, cfg *config, r *http.Request, obj any) (any, error) {
	metadata.GetMetadata(obj).ID = uuid.NewString()

	obj, err := cfg.checkWrite(obj, nil, api, r)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "write check failed (%w)", err)
	}

	err = api.sb.Write(ctx, cfg.typeName, obj)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "write failed (%w)", err)
	}

	obj, err = cfg.checkRead(obj, api, r)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "read check failed (%w)", err)
	}

	return obj, nil
}

func (api *API) getInt(ctx context.Context, cfg *config, r *http.Request, id string) (any, error) {
	obj, err := api.sb.Read(ctx, cfg.typeName, id, cfg.factory)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	if obj == nil {
		return nil, nil
	}

	obj, err = cfg.checkRead(obj, api, r)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "read check failed (%w)", err)
	}

	return obj, nil
}

func (api *API) listInt(ctx context.Context, cfg *config, r *http.Request, opts *ListOpts) ([]any, error) {
	// TODO: Add query condition pushdown

	if opts == nil {
		opts = &ListOpts{}
	}

	list, err := api.sb.List(ctx, cfg.typeName, cfg.factory)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "read list failed (%w)", err)
	}

	list, err = api.filterList(cfg, r, opts, list)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "filter list failed (%w)", err)
	}

	return list, nil
}

func (api *API) replaceInt(ctx context.Context, cfg *config, r *http.Request, id string, replace any) (any, error) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	obj, err := api.sb.Read(r.Context(), cfg.typeName, id, cfg.factory)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	if obj == nil {
		return nil, jsrest.Errorf(jsrest.ErrNotFound, "%s", id)
	}

	err = ifMatch(obj, r)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "match failed (%w)", err)
	}

	prev, err := cfg.clone(obj)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "clone failed (%w)", err)
	}

	// Metadata is immutable or server-owned
	metadata.ClearMetadata(replace)
	objMD := metadata.GetMetadata(obj)
	replaceMD := metadata.GetMetadata(replace)
	replaceMD.ID = id
	replaceMD.Generation = objMD.Generation + 1

	replace, err = cfg.checkWrite(replace, prev, api, r)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "write check failed (%w)", err)
	}

	err = api.sb.Write(r.Context(), cfg.typeName, replace)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "write failed: %s (%w)", id, err)
	}

	replace, err = cfg.checkRead(replace, api, r)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "read check failed (%w)", err)
	}

	return replace, nil
}

func (api *API) updateInt(ctx context.Context, cfg *config, r *http.Request, id string, patch any) (any, error) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	// Metadata is immutable or server-owned
	metadata.ClearMetadata(patch)

	obj, err := api.sb.Read(ctx, cfg.typeName, id, cfg.factory)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	if obj == nil {
		return nil, jsrest.Errorf(jsrest.ErrNotFound, "%s", id)
	}

	err = ifMatch(obj, r)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "match failed (%w)", err)
	}

	prev, err := cfg.clone(obj)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "clone failed (%w)", err)
	}

	merge(obj, patch)
	metadata.GetMetadata(obj).Generation++

	obj, err = cfg.checkWrite(obj, prev, api, r)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "write check failed (%w)", err)
	}

	err = api.sb.Write(ctx, cfg.typeName, obj)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "write failed: %s (%w)", id, err)
	}

	obj, err = cfg.checkRead(obj, api, r)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "read check failed (%w)", err)
	}

	return obj, nil
}
