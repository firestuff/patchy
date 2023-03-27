package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
	"github.com/firestuff/patchy/path"
)

type ListOpts struct {
	Stream  string
	Limit   int64
	Offset  int64
	After   string
	Sorts   []string
	Filters []*Filter

	IfNoneMatchETag string

	Prev any
}

type Filter struct {
	Path  string
	Op    string
	Value string
}

var (
	opMatch     = regexp.MustCompile(`^([^\[]+)\[(.+)\]$`)
	validStream = map[string]bool{
		"full": true,
		"diff": true,
	}
	validOps = map[string]bool{
		"eq":  true,
		"gt":  true,
		"gte": true,
		"hp":  true,
		"in":  true,
		"lt":  true,
		"lte": true,
	}
	ErrInvalidFilterOp     = errors.New("invalid filter operator")
	ErrInvalidSort         = errors.New("invalid _sort")
	ErrInvalidStreamFormat = errors.New("invalid _stream")
)

func ApplySorts[T any](list []T, opts *ListOpts) ([]T, error) {
	for _, srt := range opts.Sorts {
		switch {
		case strings.HasPrefix(srt, "+"):
			err := path.Sort(list, strings.TrimPrefix(srt, "+"))
			if err != nil {
				return nil, err
			}

		case strings.HasPrefix(srt, "-"):
			err := path.SortReverse(list, strings.TrimPrefix(srt, "-"))
			if err != nil {
				return nil, err
			}

		default:
			err := path.Sort(list, srt)
			if err != nil {
				return nil, err
			}
		}
	}

	return list, nil
}

func ApplyFilters[T any](list []T, opts *ListOpts) ([]T, error) {
	ret := []T{}

	for _, obj := range list {
		isMatch, err := match(obj, opts.Filters)
		if err != nil {
			return nil, jsrest.Errorf(jsrest.ErrBadRequest, "match failed (%w)", err)
		}

		if isMatch {
			ret = append(ret, obj)
		}
	}

	return ret, nil
}

func ApplyWindow[T any](list []T, opts *ListOpts) ([]T, error) {
	ret := []T{}

	after := opts.After
	offset := opts.Offset
	limit := opts.Limit

	if limit == 0 {
		limit = math.MaxInt64
	}

	for _, obj := range list {
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

func HashList(list any) (string, error) {
	hash := sha256.New()
	enc := json.NewEncoder(hash)

	if err := enc.Encode(list); err != nil {
		return "", jsrest.Errorf(jsrest.ErrInternalServerError, "JSON encode failed (%w)", err)
	}

	return fmt.Sprintf("etag:%x", hash.Sum(nil)), nil
}

func parseListOpts(r *http.Request) (*ListOpts, error) {
	ret := &ListOpts{
		Sorts:   []string{},
		Filters: []*Filter{},
	}

	// TODO: Support list
	ifNoneMatch := r.Header.Get("If-None-Match")

	if ifNoneMatch != "" {
		val, err := trimQuotes(ifNoneMatch)
		if err != nil {
			return nil, jsrest.Errorf(jsrest.ErrBadRequest, "trim quotes failed (%w) (%w)", err, ErrInvalidIfNoneMatch)
		}

		switch {
		case strings.HasPrefix(val, "etag:"):
			ret.IfNoneMatchETag = val

		default:
			return nil, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", ifNoneMatch, ErrIfNoneMatchUnknownType)
		}
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrBadRequest, "parse URL query failed (%w)", err)
	}

	if stream := params.Get("_stream"); stream != "" {
		if _, valid := validStream[stream]; !valid {
			return nil, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", stream, ErrInvalidStreamFormat)
		}

		ret.Stream = stream

		params.Del("_stream")
	}

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
	list, err := cfg.checkReadList(ctx, list, api)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "check read list failed (%w)", err)
	}

	list, err = ApplyFilters(list, opts)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrBadRequest, "filter failed (%w)", err)
	}

	list, err = ApplySorts(list, opts)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrBadRequest, "sort failed (%w)", err)
	}

	list, err = ApplyWindow(list, opts)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrBadRequest, "window failed (%w)", err)
	}

	return list, nil
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
