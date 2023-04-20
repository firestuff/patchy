package patchyc

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gopatchy/jsrest"
	"golang.org/x/exp/slices"
)

var (
	ErrNotFound            = fmt.Errorf("not found")
	ErrMultipleFound       = fmt.Errorf("multiple found")
	ErrInvalidStreamEvent  = fmt.Errorf("invalid stream event")
	ErrInvalidStreamFormat = fmt.Errorf("invalid stream format")
)

func CreateName[TOut, TIn any](ctx context.Context, c *Client, name string, obj *TIn) (*TOut, error) {
	created := new(TOut)

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

func Create[TOut, TIn any](ctx context.Context, c *Client, obj *TIn) (*TOut, error) {
	return CreateName[TOut, TIn](ctx, c, objName(new(TOut)), obj)
}

func DeleteName[TOut any](ctx context.Context, c *Client, name, id string, opts *UpdateOpts) error {
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

func Delete[TOut any](ctx context.Context, c *Client, id string, opts *UpdateOpts) error {
	return DeleteName[TOut](ctx, c, objName(new(TOut)), id, opts)
}

func FindName[TOut any](ctx context.Context, c *Client, name, shortID string) (*TOut, error) {
	listOpts := &ListOpts{
		Filters: []*Filter{
			{
				Path:  "id",
				Op:    "hp",
				Value: shortID,
			},
		},
	}

	objs, err := ListName[TOut](ctx, c, name, listOpts)
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

func Find[TOut any](ctx context.Context, c *Client, shortID string) (*TOut, error) {
	return FindName[TOut](ctx, c, objName(new(TOut)), shortID)
}

func GetName[TOut any](ctx context.Context, c *Client, name, id string, opts *GetOpts) (*TOut, error) {
	obj := new(TOut)

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
		return opts.Prev.(*TOut), nil
	}

	if resp.IsError() {
		return nil, jsrest.ReadError(resp.Body())
	}

	return obj, nil
}

func Get[TOut any](ctx context.Context, c *Client, id string, opts *GetOpts) (*TOut, error) {
	return GetName[TOut](ctx, c, objName(new(TOut)), id, opts)
}

func ListName[TOut any](ctx context.Context, c *Client, name string, opts *ListOpts) ([]*TOut, error) {
	objs := []*TOut{}

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
		return opts.Prev.([]*TOut), nil
	}

	if resp.IsError() {
		return nil, jsrest.ReadError(resp.Body())
	}

	return objs, nil
}

func List[TOut any](ctx context.Context, c *Client, opts *ListOpts) ([]*TOut, error) {
	return ListName[TOut](ctx, c, objName(new(TOut)), opts)
}

func ReplaceName[TOut, TIn any](ctx context.Context, c *Client, name, id string, obj *TIn, opts *UpdateOpts) (*TOut, error) {
	replaced := new(TOut)

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

func Replace[TOut, TIn any](ctx context.Context, c *Client, id string, obj *TIn, opts *UpdateOpts) (*TOut, error) {
	return ReplaceName[TOut, TIn](ctx, c, objName(new(TOut)), id, obj, opts)
}

func UpdateName[TOut, TIn any](ctx context.Context, c *Client, name, id string, obj *TIn, opts *UpdateOpts) (*TOut, error) {
	updated := new(TOut)

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

func Update[TOut, TIn any](ctx context.Context, c *Client, id string, obj *TIn, opts *UpdateOpts) (*TOut, error) {
	return UpdateName[TOut, TIn](ctx, c, objName(new(TOut)), id, obj, opts)
}

func StreamGetName[TOut any](ctx context.Context, c *Client, name, id string, opts *GetOpts) (*GetStream[TOut], error) {
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

	stream := &GetStream[TOut]{
		ch:   make(chan *TOut, 100),
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
				obj := new(TOut)

				err = event.decode(obj)
				if err != nil {
					stream.writeError(err)
					return
				}

				stream.writeEvent(obj)

			case "notModified":
				if opts != nil && opts.Prev != nil {
					stream.writeEvent(opts.Prev.(*TOut))
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

func StreamGet[TOut any](ctx context.Context, c *Client, id string, opts *GetOpts) (*GetStream[TOut], error) {
	return StreamGetName[TOut](ctx, c, objName(new(TOut)), id, opts)
}

func StreamListName[TOut any](ctx context.Context, c *Client, name string, opts *ListOpts) (*ListStream[TOut], error) {
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

	stream := &ListStream[TOut]{
		ch:   make(chan []*TOut, 100),
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

func streamListFull[TOut any](scan *bufio.Scanner, stream *ListStream[TOut], opts *ListOpts) {
	for {
		event, err := readEvent(scan)
		if err != nil {
			stream.writeError(err)
			return
		}

		switch event.eventType {
		case "list":
			list := []*TOut{}

			err = event.decode(&list)
			if err != nil {
				stream.writeError(err)
				return
			}

			stream.writeEvent(list)

		case "notModified":
			if opts != nil && opts.Prev != nil {
				stream.writeEvent(opts.Prev.([]*TOut))
			} else {
				stream.writeError(fmt.Errorf("notModified without If-None-Match (%w)", ErrInvalidStreamEvent))
				return
			}

		case "heartbeat":
			stream.writeHeartbeat()
		}
	}
}

func streamListDiff[TOut any](scan *bufio.Scanner, stream *ListStream[TOut], opts *ListOpts) {
	list := []*TOut{}

	add := func(event *streamEvent) error {
		obj := new(TOut)

		err := event.decode(obj)
		if err != nil {
			return err
		}

		pos, err := strconv.Atoi(event.params["new-position"])
		if err != nil {
			return err
		}

		list = slices.Insert(list, pos, obj)

		return nil
	}

	remove := func(event *streamEvent) error {
		pos, err := strconv.Atoi(event.params["old-position"])
		if err != nil {
			return err
		}

		list = slices.Delete(list, pos, pos+1)

		return nil
	}

	for {
		event, err := readEvent(scan)
		if err != nil {
			stream.writeError(err)
			return
		}

		switch event.eventType {
		case "add":
			err = add(event)
			if err != nil {
				stream.writeError(err)
				return
			}

		case "update":
			err = remove(event)
			if err != nil {
				stream.writeError(err)
				return
			}

			err = add(event)
			if err != nil {
				stream.writeError(err)
				return
			}

		case "remove":
			err = remove(event)
			if err != nil {
				stream.writeError(err)
				return
			}

		case "sync":
			stream.writeEvent(list)

		case "notModified":
			list = opts.Prev.([]*TOut)
			stream.writeEvent(list)

		case "heartbeat":
			stream.writeHeartbeat()
		}
	}
}

func StreamList[TOut any](ctx context.Context, c *Client, opts *ListOpts) (*ListStream[TOut], error) {
	return StreamListName[TOut](ctx, c, objName(new(TOut)), opts)
}

func objName[T any](obj *T) string {
	return strings.ToLower(reflect.TypeOf(*obj).Name())
}
