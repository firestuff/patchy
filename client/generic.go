package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func CreateName[T any](ctx context.Context, c *Client, name string, obj *T) (*T, error) {
	created := new(T)

	resp, err := c.rst.R().
		SetContext(ctx).
		SetPathParam("name", name).
		SetBody(obj).
		SetResult(created).
		Post("{name}")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, errors.New(resp.String())
	}

	return created, nil
}

func Create[T any](ctx context.Context, c *Client, obj *T) (*T, error) {
	return CreateName[T](ctx, c, objName(obj), obj)
}

func FindName[T any](ctx context.Context, c *Client, name, shortID string) (*T, error) {
	listOpts := &ListOpts{
		Filters: []*Filter{
			{
				Path:  "id",
				Op:    "hp",
				Value: shortID,
			},
		},
	}

	objs, err := ListName[T](ctx, c, name, listOpts)
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

func Find[T any](ctx context.Context, c *Client, shortID string) (*T, error) {
	return FindName[T](ctx, c, objName(new(T)), shortID)
}

func GetName[T any](ctx context.Context, c *Client, name, id string) (*T, error) {
	obj := new(T)

	resp, err := c.rst.R().
		SetContext(ctx).
		SetPathParam("name", name).
		SetPathParam("id", id).
		SetResult(obj).
		Get("{name}/{id}")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, errors.New(resp.String())
	}

	return obj, nil
}

func Get[T any](ctx context.Context, c *Client, id string) (*T, error) {
	return GetName[T](ctx, c, objName(new(T)), id)
}

func ListName[T any](ctx context.Context, c *Client, name string, opts *ListOpts) ([]*T, error) {
	objs := []*T{}

	r := c.rst.R().
		SetContext(ctx).
		SetPathParam("name", name).
		SetResult(&objs)

	if opts != nil {
		applyListOpts(opts, r)
	}

	resp, err := r.Get("{name}")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, errors.New(resp.String())
	}

	return objs, nil
}

func List[T any](ctx context.Context, c *Client, opts *ListOpts) ([]*T, error) {
	return ListName[T](ctx, c, objName(new(T)), opts)
}

func ReplaceName[T any](ctx context.Context, c *Client, name, id string, obj *T) (*T, error) {
	replaced := new(T)

	resp, err := c.rst.R().
		SetContext(ctx).
		SetPathParam("name", name).
		SetPathParam("id", id).
		SetBody(obj).
		SetResult(replaced).
		Put("{name}/{id}")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, errors.New(resp.String())
	}

	return replaced, nil
}

func Replace[T any](ctx context.Context, c *Client, id string, obj *T) (*T, error) {
	return ReplaceName[T](ctx, c, objName(obj), id, obj)
}

func UpdateName[T any](ctx context.Context, c *Client, name, id string, obj *T) (*T, error) {
	updated := new(T)

	resp, err := c.rst.R().
		SetContext(ctx).
		SetPathParam("name", name).
		SetPathParam("id", id).
		SetBody(obj).
		SetResult(updated).
		Patch("{name}/{id}")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, errors.New(resp.String())
	}

	return updated, nil
}

func Update[T any](ctx context.Context, c *Client, id string, obj *T) (*T, error) {
	return UpdateName[T](ctx, c, objName(obj), id, obj)
}

func objName[T any](obj *T) string {
	return strings.ToLower(reflect.TypeOf(*obj).Name())
}
