package patchy

import "math"
import "net/http"
import "net/url"
import "strconv"

import "github.com/firestuff/patchy/metadata"
import "github.com/firestuff/patchy/path"

type listParams struct {
	limit   int64
	offset  int64
	after   string
	filters url.Values
}

func parseListParams(params url.Values) (*listParams, error) {
	ret := &listParams{
		limit:  math.MaxInt64,
		offset: 0,
		after:  "",
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

	ret.filters = params

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

	return ret, nil
}

func match(obj any, params url.Values) (bool, error) {
	for fieldPath, vals := range params {
		for _, val := range vals {
			matches, err := path.Match(obj, fieldPath, val)
			if err != nil {
				return false, err
			}
			if !matches {
				return false, nil
			}
		}
	}

	return true, nil
}
