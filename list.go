package patchy

import "fmt"
import "math"
import "net/http"
import "net/url"
import "regexp"
import "strconv"
import "strings"

import "github.com/firestuff/patchy/metadata"
import "github.com/firestuff/patchy/path"

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

var opMatch = regexp.MustCompile(`^([^\[]+)\[(.+)\]$`)
var validOps = map[string]bool{
	"eq": true,
	"gt": true,
}

func parseListParams(params url.Values) (*listParams, error) {
	ret := &listParams{
		limit:   math.MaxInt64,
		sorts:   []string{},
		filters: []filter{},
	}

	var err error

	limit := params.Get("_limit")
	if limit != "" {
		ret.limit, err = strconv.ParseInt(limit, 10, 64)
		if err != nil {
			return nil, err
		}
		params.Del("_limit")
	}

	offset := params.Get("_offset")
	if offset != "" {
		ret.offset, err = strconv.ParseInt(offset, 10, 64)
		if err != nil {
			return nil, err
		}
		params.Del("_offset")
	}

	ret.after = params.Get("_after")
	if ret.after != "" {
		params.Del("_after")
	}

	sorts := params["_sort"]
	for i := len(sorts) - 1; i >= 0; i-- {
		srt := sorts[i]
		if len(srt) == 0 {
			return nil, fmt.Errorf("invalid _sort: %s", srt)
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
				return nil, fmt.Errorf("invalid filter operator: %s", f.op)
			}

			ret.filters = append(ret.filters, f)
		}
	}

	return ret, nil
}

func (api *API) list(cfg *config, r *http.Request, params *listParams) ([]any, error) {
	list, err := api.sb.List(cfg.typeName, cfg.factory)
	if err != nil {
		return nil, err
	}

	ret := []any{}

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

		if params.after != "" {
			if metadata.GetMetadata(obj).Id == params.after {
				params.after = ""
			}
			continue
		}

		if params.offset > 0 {
			params.offset--
			continue
		}

		params.limit--
		if params.limit < 0 {
			break
		}

		ret = append(ret, obj)
	}

	for _, srt := range params.sorts {
		switch {
		case strings.HasPrefix(srt, "+"):
			err = path.Sort(ret, strings.TrimPrefix(srt, "+"))

		case strings.HasPrefix(srt, "-"):
			err = path.SortReverse(ret, strings.TrimPrefix(srt, "-"))

		default:
			err = path.Sort(ret, srt)
		}
		if err != nil {
			return nil, err
		}
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
