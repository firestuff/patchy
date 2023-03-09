package patchy

import (
	"errors"
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
		"hp":  true,
		"in":  true,
		"lt":  true,
		"lte": true,
	}
	ErrInvalidFilterOp = errors.New("invalid filter operator")
	ErrInvalidSort     = errors.New("invalid _sort")
)

func parseListParams(params url.Values) (*listParams, error) {
	ret := &listParams{
		limit:   math.MaxInt64,
		sorts:   []string{},
		filters: []filter{},
	}

	var err error

	if limit := params.Get("_limit"); limit != "" {
		ret.limit, err = strconv.ParseInt(limit, 10, 64)
		if err != nil {
			return nil, jsrest.Errorf(jsrest.ErrBadRequest, "parse _limit value failed: %s (%w)", limit, err)
		}

		params.Del("_limit")
	}

	if offset := params.Get("_offset"); offset != "" {
		ret.offset, err = strconv.ParseInt(offset, 10, 64)
		if err != nil {
			return nil, jsrest.Errorf(jsrest.ErrBadRequest, "parse _offset value failed: %s (%w)", offset, err)
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
			return nil, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", srt, ErrInvalidSort)
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
				return nil, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", f.op, ErrInvalidFilterOp)
			}

			ret.filters = append(ret.filters, f)
		}
	}

	return ret, nil
}

func filterList(cfg *config, r *http.Request, params *listParams, list []any) ([]any, error) {
	inter := []any{}

	for _, obj := range list {
		obj, jse := cfg.checkRead(obj, r)
		if jse != nil {
			continue
		}

		matches, err := match(obj, params.filters)
		if err != nil {
			return nil, jsrest.Errorf(jsrest.ErrBadRequest, "match failed (%w)", err)
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
			return nil, jsrest.Errorf(jsrest.ErrBadRequest, "sort failed (%w)", err)
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

		case "hp":
			matches, err = path.HasPrefix(obj, filter.path, filter.val)

		case "in":
			matches, err = path.In(obj, filter.path, filter.val)

		case "lt":
			matches, err = path.Less(obj, filter.path, filter.val)

		case "lte":
			matches, err = path.LessEqual(obj, filter.path, filter.val)

		default:
			panic(filter.op)
		}

		if err != nil {
			return false, jsrest.Errorf(jsrest.ErrBadRequest, "match operation failed: %s[%s] (%w)", filter.path, filter.op, err)
		}

		if !matches {
			return false, nil
		}
	}

	return true, nil
}
