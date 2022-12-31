package patchy

import (
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
	"github.com/firestuff/patchy/view"
)

type listParams struct {
	limit   int64
	offset  int64
	after   string
	sorts   []string
	filters []filter
}

type filter struct {
	path string
	op   string
	val  string
}

var (
	opMatch  = regexp.MustCompile(`^([^\[]+)\[(.+)\]$`)
	validOps = map[string]bool{
		"eq":  true,
		"gt":  true,
		"gte": true,
		"in":  true,
		"lt":  true,
		"lte": true,
	}
	ErrInvalidFilterOp = errors.New("invalid filter operator")
	ErrInvalidSort     = errors.New("invalid _sort")
)

func parseListParams(params url.Values) (*listParams, *jsrest.Error) {
	ret := &listParams{
		limit:   math.MaxInt64,
		sorts:   []string{},
		filters: []filter{},
	}

	var err error

	if limit := params.Get("_limit"); limit != "" {
		ret.limit, err = strconv.ParseInt(limit, 10, 64)
		if err != nil {
			e := fmt.Errorf("failed to parse _limit value %s: %w", limit, err)
			jse := jsrest.FromError(e, jsrest.StatusBadRequest)

			return nil, jse
		}

		params.Del("_limit")
	}

	if offset := params.Get("_offset"); offset != "" {
		ret.offset, err = strconv.ParseInt(offset, 10, 64)
		if err != nil {
			e := fmt.Errorf("failed to parse _offset value %s: %w", offset, err)
			jse := jsrest.FromError(e, jsrest.StatusBadRequest)

			return nil, jse
		}

		params.Del("_offset")
	}

	if ret.after = params.Get("_after"); ret.after != "" {
		params.Del("_after")
	}

	sorts := params["_sort"]
	for i := len(sorts) - 1; i >= 0; i-- {
		srt := sorts[i]
		if len(srt) == 0 {
			e := fmt.Errorf("%s: %w", srt, ErrInvalidSort)
			jse := jsrest.FromError(e, jsrest.StatusBadRequest)

			return nil, jse
		}

		ret.sorts = append(ret.sorts, srt)
	}
	params.Del("_sort")

	for path, vals := range params {
		for _, val := range vals {
			f := filter{
				path: path,
				op:   "eq",
				val:  val,
			}

			matches := opMatch.FindStringSubmatch(f.path)
			if matches != nil {
				f.path = matches[1]
				f.op = matches[2]
			}

			if _, valid := validOps[f.op]; !valid {
				e := fmt.Errorf("%s: %w", f.op, ErrInvalidFilterOp)
				jse := jsrest.FromError(e, jsrest.StatusBadRequest)

				return nil, jse
			}

			ret.filters = append(ret.filters, f)
		}
	}

	return ret, nil
}

func (api *API) list(cfg *config, r *http.Request, params *listParams) (view.ReadView[[]any], *jsrest.Error) {
	v, err := api.sb.List(r.Context(), cfg.typeName, cfg.factory)
	if err != nil {
		e := fmt.Errorf("failed to read list: %w", err)
		jse := jsrest.FromError(e, jsrest.StatusInternalServerError)

		return nil, jse
	}

	return filterList(cfg, r, params, v), nil
}

func filterList(cfg *config, r *http.Request, params *listParams, in view.ReadView[[]any]) *view.FilterView[[]any] {
	// TODO: Stack FilterViews instead of using one giant one
	return view.NewFilterView[[]any](in, func(list []any) ([]any, error) {
		inter := []any{}

		for _, obj := range list {
			if cfg.mayRead != nil {
				if cfg.mayRead(obj, r) != nil {
					continue
				}
			}

			matches, err := match(obj, params.filters)
			if err != nil {
				return nil, err
			}

			if !matches {
				continue
			}

			inter = append(inter, obj)
		}

		for _, srt := range params.sorts {
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
				return nil, err
			}
		}

		ret := []any{}

		after := params.after
		offset := params.offset
		limit := params.limit

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
	})
}

func match(obj any, filters []filter) (bool, error) {
	for _, filter := range filters {
		var matches bool

		var err error

		switch filter.op {
		case "eq":
			matches, err = path.Equal(obj, filter.path, filter.val)

		case "gt":
			matches, err = path.Greater(obj, filter.path, filter.val)

		case "gte":
			matches, err = path.GreaterEqual(obj, filter.path, filter.val)

		case "in":
			matches, err = path.In(obj, filter.path, filter.val)

		case "lt":
			matches, err = path.Less(obj, filter.path, filter.val)

		case "lte":
			matches, err = path.LessEqual(obj, filter.path, filter.val)
		}

		if err != nil {
			return false, err
		}

		if !matches {
			return false, nil
		}
	}

	return true, nil
}
