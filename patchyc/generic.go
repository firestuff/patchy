package patchyc

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/firestuff/patchy/api"
	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

var (
	ErrNotFound            = fmt.Errorf("not found")
	ErrMultipleFound       = fmt.Errorf("multiple found")
	ErrInvalidStreamEvent  = fmt.Errorf("invalid stream event")
	ErrInvalidStreamFormat = fmt.Errorf("invalid stream format")
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

func DeleteName[T any](ctx context.Context, c *Client, name, id string, opts *UpdateOpts) error {
	r := c.rst.R().
		SetContext(ctx).
		SetPathParam("name", name).
		SetPathParam("id", id)

	if opts != nil {
		applyUpdateOpts(opts, r)
	}

	resp, err := r.Delete("{name}/{id}")
	if err != nil {
		return err
	}

	if resp.IsError() {
		return jsrest.ReadError(resp.Body())
	}

	return nil
}

func Delete[T any](ctx context.Context, c *Client, id string, opts *UpdateOpts) error {
	return DeleteName[T](ctx, c, objName(new(T)), id, opts)
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

func GetName[T any](ctx context.Context, c *Client, name, id string, opts *GetOpts) (*T, error) {
	obj := new(T)

	r := c.rst.R().
		SetContext(ctx).
		SetPathParam("name", name).
		SetPathParam("id", id).
		SetResult(obj)

	if opts != nil {
		applyGetOpts(opts, r)
	}

	resp, err := r.Get("{name}/{id}")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == http.StatusNotFound {
		return nil, nil
	}

	if opts != nil && opts.Prev != nil && resp.StatusCode() == http.StatusNotModified {
		return opts.Prev.(*T), nil
	}

	if resp.IsError() {
		return nil, jsrest.ReadError(resp.Body())
	}

	return obj, nil
}

func Get[T any](ctx context.Context, c *Client, id string, opts *GetOpts) (*T, error) {
	return GetName[T](ctx, c, objName(new(T)), id, opts)
}

func ListName[T any](ctx context.Context, c *Client, name string, opts *ListOpts) ([]*T, error) {
	objs := []*T{}

	r := c.rst.R().
		SetContext(ctx).
		SetPathParam("name", name).
		SetResult(&objs)

	if opts != nil {
		err := applyListOpts(opts, r)
		if err != nil {
			return nil, err
		}
	}

	resp, err := r.Get("{name}")
	if err != nil {
		return nil, err
	}

	if opts != nil && opts.Prev != nil && resp.StatusCode() == http.StatusNotModified {
		return opts.Prev.([]*T), nil
	}

	if resp.IsError() {
		return nil, jsrest.ReadError(resp.Body())
	}

	return objs, nil
}

func List[T any](ctx context.Context, c *Client, opts *ListOpts) ([]*T, error) {
	return ListName[T](ctx, c, objName(new(T)), opts)
}

func ReplaceName[T any](ctx context.Context, c *Client, name, id string, obj *T, opts *UpdateOpts) (*T, error) {
	replaced := new(T)

	r := c.rst.R().
		SetContext(ctx).
		SetPathParam("name", name).
		SetPathParam("id", id).
		SetBody(obj).
		SetResult(replaced)

	if opts != nil {
		applyUpdateOpts(opts, r)
	}

	resp, err := r.Put("{name}/{id}")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, jsrest.ReadError(resp.Body())
	}

	return replaced, nil
}

func Replace[T any](ctx context.Context, c *Client, id string, obj *T, opts *UpdateOpts) (*T, error) {
	return ReplaceName[T](ctx, c, objName(obj), id, obj, opts)
}

func UpdateName[T any](ctx context.Context, c *Client, name, id string, obj *T, opts *UpdateOpts) (*T, error) {
	updated := new(T)

	r := c.rst.R().
		SetContext(ctx).
		SetPathParam("name", name).
		SetPathParam("id", id).
		SetBody(obj).
		SetResult(updated)

	if opts != nil {
		applyUpdateOpts(opts, r)
	}

	resp, err := r.Patch("{name}/{id}")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, jsrest.ReadError(resp.Body())
	}

	return updated, nil
}

func Update[T any](ctx context.Context, c *Client, id string, obj *T, opts *UpdateOpts) (*T, error) {
	return UpdateName[T](ctx, c, objName(obj), id, obj, opts)
}

func StreamGetName[T any](ctx context.Context, c *Client, name, id string, opts *GetOpts) (*GetStream[T], error) {
	r := c.rst.R().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetPathParam("name", name).
		SetPathParam("id", id)

	if opts != nil {
		applyGetOpts(opts, r)
	}

	resp, err := r.Get("{name}/{id}")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, jsrest.ReadError(resp.Body())
	}

	body := resp.RawBody()
	scan := bufio.NewScanner(body)

	stream := &GetStream[T]{
		ch:   make(chan *GetStreamEvent[T], 100),
		body: body,
	}

	go func() {
		for {
			event, err := readEvent(scan)
			if err != nil {
				stream.writeError(err)
				return
			}

			switch event.eventType {
			case "initial":
				fallthrough
			case "update":
				obj := new(T)

				err = event.decode(obj)
				if err != nil {
					stream.writeError(err)
					return
				}

				stream.writeEvent(event.id, obj)

			case "notModified":
				if opts != nil && opts.Prev != nil {
					stream.writeEvent(event.id, opts.Prev.(*T))
				} else {
					stream.writeError(fmt.Errorf("notModified without If-None-Match (%w)", ErrInvalidStreamEvent))
					return
				}

			case "heartbeat":
				stream.writeHeartbeat()
			}
		}
	}()

	return stream, nil
}

func StreamGet[T any](ctx context.Context, c *Client, id string, opts *GetOpts) (*GetStream[T], error) {
	return StreamGetName[T](ctx, c, objName(new(T)), id, opts)
}

func StreamListName[T any](ctx context.Context, c *Client, name string, opts *ListOpts) (*ListStream[T], error) {
	r := c.rst.R().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetPathParam("name", name)

	if opts == nil {
		opts = &ListOpts{}
	}

	if opts != nil {
		err := applyListOpts(opts, r)
		if err != nil {
			return nil, err
		}
	}

	resp, err := r.Get("{name}")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, jsrest.ReadError(resp.Body())
	}

	body := resp.RawBody()
	scan := bufio.NewScanner(body)

	stream := &ListStream[T]{
		ch:   make(chan *ListStreamEvent[T], 100),
		body: body,
	}

	switch resp.Header().Get("Stream-Format") {
	case "full":
		go streamListFull(scan, stream, opts)

	case "diff":
		go streamListDiff(scan, stream, opts)

	default:
		stream.Close()
		return nil, fmt.Errorf("%s (%w)", resp.Header().Get("Stream-Format"), ErrInvalidStreamFormat)
	}

	return stream, nil
}

func streamListFull[T any](scan *bufio.Scanner, stream *ListStream[T], opts *ListOpts) {
	for {
		event, err := readEvent(scan)
		if err != nil {
			stream.writeError(err)
			return
		}

		switch event.eventType {
		case "list":
			list := []*T{}

			err = event.decode(&list)
			if err != nil {
				stream.writeError(err)
				return
			}

			stream.writeEvent(event.id, list)

		case "notModified":
			if opts != nil && opts.Prev != nil {
				stream.writeEvent(event.id, opts.Prev.([]*T))
			} else {
				stream.writeError(fmt.Errorf("notModified without If-None-Match (%w)", ErrInvalidStreamEvent))
				return
			}

		case "heartbeat":
			stream.writeHeartbeat()
		}
	}
}

func streamListDiff[T any](scan *bufio.Scanner, stream *ListStream[T], opts *ListOpts) {
	objs := map[string]*T{}

	for {
		event, err := readEvent(scan)
		if err != nil {
			stream.writeError(err)
			return
		}

		switch event.eventType {
		case "add":
			fallthrough
		case "update":
			obj := new(T)

			err = event.decode(obj)
			if err != nil {
				stream.writeError(err)
				return
			}

			id := metadata.GetMetadata(obj).ID
			objs[id] = obj

		case "remove":
			obj := new(T)

			err = event.decode(obj)
			if err != nil {
				stream.writeError(err)
				return
			}

			id := metadata.GetMetadata(obj).ID
			delete(objs, id)

		case "sync":
			list := []*T{}

			for _, obj := range objs {
				list = append(list, obj)
			}

			list, err := api.ApplySorts(list, opts)
			if err != nil {
				stream.writeError(err)
				return
			}

			stream.writeEvent(event.id, list)

		case "notModified":
			if opts != nil && opts.Prev != nil {
				stream.writeEvent(event.id, opts.Prev.([]*T))
			} else {
				stream.writeError(fmt.Errorf("notModified without If-None-Match (%w)", ErrInvalidStreamEvent))
				return
			}

		case "heartbeat":
			stream.writeHeartbeat()
		}
	}
}

func StreamList[T any](ctx context.Context, c *Client, opts *ListOpts) (*ListStream[T], error) {
	return StreamListName[T](ctx, c, objName(new(T)), opts)
}

func objName[T any](obj *T) string {
	return strings.ToLower(reflect.TypeOf(*obj).Name())
}
