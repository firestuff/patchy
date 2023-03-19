package api

import (
	"context"

	"github.com/dchest/uniuri"
	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

type intObjectStream struct {
	ch <-chan any

	api    *API
	cfg    *config
	id     string
	sbChan <-chan any
}

type intObjectListStream struct {
	ch <-chan []any

	api    *API
	cfg    *config
	sbChan <-chan []any
}

func (api *API) createInt(ctx context.Context, cfg *config, obj any) (any, error) {
	// TODO: Remove http.Request argument from all these functions
	metadata.GetMetadata(obj).ID = uniuri.New()

	obj, err := cfg.checkWrite(ctx, obj, nil, api)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "write check failed (%w)", err)
	}

	err = api.sb.Write(ctx, cfg.typeName, obj)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "write failed (%w)", err)
	}

	obj, err = cfg.checkRead(ctx, obj, api)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "read check failed (%w)", err)
	}

	return obj, nil
}

func (api *API) deleteInt(ctx context.Context, cfg *config, id string) error {
	obj, err := api.sb.Read(ctx, cfg.typeName, id, cfg.factory)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	if obj == nil {
		return jsrest.Errorf(jsrest.ErrNotFound, "%s", id)
	}

	_, err = cfg.checkWrite(ctx, nil, obj, api)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrUnauthorized, "write check failed (%w)", err)
	}

	err = api.sb.Delete(ctx, cfg.typeName, id)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "delete failed: %s (%w)", id, err)
	}

	return nil
}

func (api *API) getInt(ctx context.Context, cfg *config, id string) (any, error) {
	obj, err := api.sb.Read(ctx, cfg.typeName, id, cfg.factory)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	if obj == nil {
		return nil, nil
	}

	obj, err = cfg.checkRead(ctx, obj, api)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "read check failed (%w)", err)
	}

	return obj, nil
}

func (api *API) listInt(ctx context.Context, cfg *config, opts *ListOpts) ([]any, error) {
	// TODO: Add query condition pushdown

	if opts == nil {
		opts = &ListOpts{}
	}

	// TODO: Add a hook for the type to mutate opts

	list, err := api.sb.List(ctx, cfg.typeName, cfg.factory)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "read list failed (%w)", err)
	}

	list, err = api.filterList(ctx, cfg, opts, list)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "filter list failed (%w)", err)
	}

	return list, nil
}

func (api *API) replaceInt(ctx context.Context, cfg *config, ifmatch, id string, replace any) (any, error) {
	cfg.lock(id)
	defer cfg.unlock(id)

	obj, err := api.sb.Read(ctx, cfg.typeName, id, cfg.factory)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	if obj == nil {
		return nil, jsrest.Errorf(jsrest.ErrNotFound, "%s", id)
	}

	err = ifMatch(obj, ifmatch)
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

	replace, err = cfg.checkWrite(ctx, replace, prev, api)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "write check failed (%w)", err)
	}

	err = api.sb.Write(ctx, cfg.typeName, replace)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "write failed: %s (%w)", id, err)
	}

	replace, err = cfg.checkRead(ctx, replace, api)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "read check failed (%w)", err)
	}

	return replace, nil
}

func (api *API) updateInt(ctx context.Context, cfg *config, ifmatch, id string, patch any) (any, error) {
	cfg.lock(id)
	defer cfg.unlock(id)

	// Metadata is immutable or server-owned
	metadata.ClearMetadata(patch)

	obj, err := api.sb.Read(ctx, cfg.typeName, id, cfg.factory)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	if obj == nil {
		return nil, jsrest.Errorf(jsrest.ErrNotFound, "%s", id)
	}

	err = ifMatch(obj, ifmatch)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "match failed (%w)", err)
	}

	prev, err := cfg.clone(obj)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "clone failed (%w)", err)
	}

	merge(obj, patch)
	metadata.GetMetadata(obj).Generation++

	obj, err = cfg.checkWrite(ctx, obj, prev, api)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "write check failed (%w)", err)
	}

	err = api.sb.Write(ctx, cfg.typeName, obj)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "write failed: %s (%w)", id, err)
	}

	obj, err = cfg.checkRead(ctx, obj, api)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "read check failed (%w)", err)
	}

	return obj, nil
}

func (api *API) getStreamInt(ctx context.Context, cfg *config, id string) (*intObjectStream, error) {
	in, err := api.sb.ReadStream(ctx, cfg.typeName, id, cfg.factory)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "read failed: %s (%w)", id, err)
	}

	// Pull the first item out of the channel and convert issues with it to errors

	obj := <-in
	if obj == nil {
		api.sb.CloseReadStream(cfg.typeName, id, in)
		return nil, jsrest.Errorf(jsrest.ErrNotFound, "%s", id)
	}

	obj, err = cfg.checkRead(ctx, obj, api)
	if err != nil {
		api.sb.CloseReadStream(cfg.typeName, id, in)
		return nil, jsrest.Errorf(jsrest.ErrUnauthorized, "read check failed (%w)", err)
	}

	out := make(chan any, 100)
	out <- obj

	go func() {
		defer close(out)

		for obj := range in {
			obj, err = cfg.checkRead(ctx, obj, api)
			if err != nil {
				break
			}

			out <- obj
		}
	}()

	return &intObjectStream{
		ch:     out,
		api:    api,
		cfg:    cfg,
		id:     id,
		sbChan: in,
	}, nil
}

func (api *API) listStreamInt(ctx context.Context, cfg *config, opts *ListOpts) (*intObjectListStream, error) {
	in, err := api.sb.ListStream(ctx, cfg.typeName, cfg.factory)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "read list failed (%w)", err)
	}

	out := make(chan []any, 100)

	go func() {
		defer close(out)

		for list := range in {
			list, err = api.filterList(ctx, cfg, opts, list)
			if err != nil {
				break
			}

			out <- list
		}
	}()

	return &intObjectListStream{
		ch:     out,
		api:    api,
		cfg:    cfg,
		sbChan: in,
	}, nil
}

func (ios *intObjectStream) Close() {
	ios.api.sb.CloseReadStream(ios.cfg.typeName, ios.id, ios.sbChan)
}

func (ios *intObjectStream) Chan() <-chan any {
	return ios.ch
}

func (iols *intObjectListStream) Close() {
	iols.api.sb.CloseListStream(iols.cfg.typeName, iols.sbChan)
}

func (iols *intObjectListStream) Chan() <-chan []any {
	return iols.ch
}
