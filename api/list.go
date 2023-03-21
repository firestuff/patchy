package api

import (
	"context"
	"errors"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
	"github.com/firestuff/patchy/path"
)

type ListOpts struct {
	// TODO: Add mode or similar (full/diff)
	Limit   int64
	Offset  int64
	After   string
	Sorts   []string
	Filters []*Filter
}

type Filter struct {
	Path  string
	Op    string
	Value string
}

var (
	opMatch  = regexp.MustCompile(`^([^\[]+)\[(.+)\]$`)
	validOps = map[string]bool{
		"eq":  true,
		"gt":  true,
		"gte": true,
		"hp":  true,
		"in":  true,
		"lt":  true,
		"lte": true,
	}
	ErrInvalidFilterOp = errors.New("invalid filter operator")
	ErrInvalidSort     = errors.New("invalid _sort")
)

func parseListOpts(params url.Values) (*ListOpts, error) {
	ret := &ListOpts{
		Sorts:   []string{},
		Filters: []*Filter{},
	}

	var err error

	if limit := params.Get("_limit"); limit != "" {
		ret.Limit, err = strconv.ParseInt(limit, 10, 64)
		if err != nil {
			return nil, jsrest.Errorf(jsrest.ErrBadRequest, "parse _limit value failed: %s (%w)", limit, err)
		}

		params.Del("_limit")
	}

	if offset := params.Get("_offset"); offset != "" {
		ret.Offset, err = strconv.ParseInt(offset, 10, 64)
		if err != nil {
			return nil, jsrest.Errorf(jsrest.ErrBadRequest, "parse _offset value failed: %s (%w)", offset, err)
		}

		params.Del("_offset")
	}

	if ret.After = params.Get("_after"); ret.After != "" {
		params.Del("_after")
	}

	sorts := params["_sort"]
	for i := len(sorts) - 1; i >= 0; i-- {
		srt := sorts[i]
		if len(srt) == 0 {
			return nil, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", srt, ErrInvalidSort)
		}

		ret.Sorts = append(ret.Sorts, srt)
	}

	params.Del("_sort")

	for path, vals := range params {
		for _, val := range vals {
			f := &Filter{
				Path:  path,
				Op:    "eq",
				Value: val,
			}

			matches := opMatch.FindStringSubmatch(f.Path)
			if matches != nil {
				f.Path = matches[1]
				f.Op = matches[2]
			}

			if _, valid := validOps[f.Op]; !valid {
				return nil, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", f.Op, ErrInvalidFilterOp)
			}

			ret.Filters = append(ret.Filters, f)
		}
	}

	return ret, nil
}

func (api *API) filterList(ctx context.Context, cfg *config, opts *ListOpts, list []any) ([]any, error) {
	inter := []any{}

	for _, obj := range list {
		obj, jse := cfg.checkRead(ctx, obj, api)
		if jse != nil {
			continue
		}

		matches, err := match(obj, opts.Filters)
		if err != nil {
			return nil, jsrest.Errorf(jsrest.ErrBadRequest, "match failed (%w)", err)
		}

		if !matches {
			continue
		}

		inter = append(inter, obj)
	}

	for _, srt := range opts.Sorts {
		var err error

		switch {
		case strings.HasPrefix(srt, "+"):
			err = path.Sort(inter, strings.TrimPrefix(srt, "+"))

		case strings.HasPrefix(srt, "-"):
			err = path.SortReverse(inter, strings.TrimPrefix(srt, "-"))

		default:
			err = path.Sort(inter, srt)
		}

		if err != nil {
			return nil, jsrest.Errorf(jsrest.ErrBadRequest, "sort failed (%w)", err)
		}
	}

	ret := []any{}

	after := opts.After
	offset := opts.Offset
	limit := opts.Limit

	if limit == 0 {
		limit = math.MaxInt64
	}

	for _, obj := range inter {
		if after != "" {
			if metadata.GetMetadata(obj).ID == after {
				after = ""
			}

			continue
		}

		if offset > 0 {
			offset--

			continue
		}

		limit--
		if limit < 0 {
			break
		}

		ret = append(ret, obj)
	}

	return ret, nil
}

func match(obj any, filters []*Filter) (bool, error) {
	for _, filter := range filters {
		var matches bool

		var err error

		switch filter.Op {
		case "eq":
			matches, err = path.Equal(obj, filter.Path, filter.Value)

		case "gt":
			matches, err = path.Greater(obj, filter.Path, filter.Value)

		case "gte":
			matches, err = path.GreaterEqual(obj, filter.Path, filter.Value)

		case "hp":
			matches, err = path.HasPrefix(obj, filter.Path, filter.Value)

		case "in":
			matches, err = path.In(obj, filter.Path, filter.Value)

		case "lt":
			matches, err = path.Less(obj, filter.Path, filter.Value)

		case "lte":
			matches, err = path.LessEqual(obj, filter.Path, filter.Value)

		default:
			return false, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", filter.Op, ErrInvalidFilterOp)
		}

		if err != nil {
			return false, jsrest.Errorf(jsrest.ErrBadRequest, "match operation failed: %s[%s] (%w)", filter.Path, filter.Op, err)
		}

		if !matches {
			return false, nil
		}
	}

	return true, nil
}
