package patchyc

import (
	"bufio"
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

var (
	ErrNotFound      = fmt.Errorf("not found")
	ErrMultipleFound = fmt.Errorf("multiple found")
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
		return nil, jsrest.ReadError(resp.Body())
	}

	return created, nil
}

func Create[T any](ctx context.Context, c *Client, obj *T) (*T, error) {
	return CreateName[T](ctx, c, objName(obj), obj)
}

func DeleteName(ctx context.Context, c *Client, name, id string) error {
	resp, err := c.rst.R().
		SetContext(ctx).
		SetPathParam("name", name).
		SetPathParam("id", id).
		Delete("{name}/{id}")
	if err != nil {
		return err
	}

	if resp.IsError() {
		return jsrest.ReadError(resp.Body())
	}

	return nil
}

func Delete[T any](ctx context.Context, c *Client, id string) error {
	return DeleteName(ctx, c, objName(new(T)), id)
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
		return nil, fmt.Errorf("%s (%w)", shortID, ErrNotFound)
	}

	if len(objs) > 1 {
		return nil, fmt.Errorf("%s (%w)", shortID, ErrMultipleFound)
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

	if resp.StatusCode() == 404 {
		return nil, nil
	}

	if resp.IsError() {
		return nil, jsrest.ReadError(resp.Body())
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
		return nil, jsrest.ReadError(resp.Body())
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
		return nil, jsrest.ReadError(resp.Body())
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
		return nil, jsrest.ReadError(resp.Body())
	}

	return updated, nil
}

func Update[T any](ctx context.Context, c *Client, id string, obj *T) (*T, error) {
	return UpdateName[T](ctx, c, objName(obj), id, obj)
}

func StreamGetName[T any](ctx context.Context, c *Client, name, id string) (*GetStream[T], error) {
	resp, err := c.rst.R().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetPathParam("name", name).
		SetPathParam("id", id).
		Get("{name}/{id}")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, jsrest.ReadError(resp.Body())
	}

	body := resp.RawBody()
	scan := bufio.NewScanner(body)

	out := make(chan *T, 100)

	go func() {
		defer close(out)

		for {
			obj := new(T)

			eventType, err := readEvent(scan, obj)
			if err != nil {
				return
			}

			if eventType == "initial" || eventType == "update" {
				out <- obj
			}
		}
	}()

	return &GetStream[T]{
		ch:   out,
		body: body,
	}, nil
}

func StreamGet[T any](ctx context.Context, c *Client, id string) (*GetStream[T], error) {
	return StreamGetName[T](ctx, c, objName(new(T)), id)
}

func StreamListName[T any](ctx context.Context, c *Client, name string, opts *ListOpts) (*ListStream[T], error) {
	r := c.rst.R().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetPathParam("name", name)

	if opts == nil {
		opts = &ListOpts{}
	}

	applyListOpts(opts, r)

	resp, err := r.Get("{name}")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, jsrest.ReadError(resp.Body())
	}

	body := resp.RawBody()
	scan := bufio.NewScanner(body)

	out := make(chan []*T, 100)

	switch opts.Stream {
	case "":
		fallthrough
	case "full":
		go streamListFull(out, scan)

	case "diff":
		go streamListDiff(out, scan, opts)
	}

	return &ListStream[T]{
		ch:   out,
		body: body,
	}, nil
}

func streamListFull[T any](out chan<- []*T, scan *bufio.Scanner) {
	defer close(out)

	for {
		list := []*T{}

		eventType, err := readEvent(scan, &list)
		if err != nil {
			return
		}

		if eventType == "list" {
			out <- list
		}
	}
}

func streamListDiff[T any](out chan<- []*T, scan *bufio.Scanner, opts *ListOpts) {
	defer close(out)

	objs := map[string]*T{}

	for {
		obj := new(T)

		eventType, err := readEvent(scan, &obj)
		if err != nil {
			return
		}

		switch eventType {
		case "add":
			id := metadata.GetMetadata(obj).ID
			objs[id] = obj

		case "update":
			id := metadata.GetMetadata(obj).ID
			objs[id] = obj

		case "remove":
			id := metadata.GetMetadata(obj).ID
			delete(objs, id)

		case "sync":
			listAny := []any{}

			for _, obj := range objs {
				listAny = append(listAny, obj)
			}

			// TODO: Make ApplySorts generic
			listAny, err := opts.ApplySorts(listAny)
			if err != nil {
				return
			}

			list := []*T{}

			for _, obj := range listAny {
				list = append(list, obj.(*T))
			}

			out <- list

		case "heartbeat":
		}
	}
}

func StreamList[T any](ctx context.Context, c *Client, opts *ListOpts) (*ListStream[T], error) {
	return StreamListName[T](ctx, c, objName(new(T)), opts)
}

func objName[T any](obj *T) string {
	return strings.ToLower(reflect.TypeOf(*obj).Name())
}
