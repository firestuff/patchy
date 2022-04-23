package api

import "math"
import "net/http"
import "net/url"
import "strconv"

import "github.com/firestuff/patchy/path"

func (api *API) list(cfg *config, r *http.Request, params url.Values) ([]any, error) {
	list, err := api.sb.List(cfg.typeName, cfg.factory)
	if err != nil {
		return nil, err
	}

	ret := []any{}

	limit := int64(math.MaxInt64)
	limitStr := params.Get("_limit")
	if limitStr != "" {
		limit, err = strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			return nil, err
		}
		params.Del("_limit")
	}

	offset := int64(0)
	offsetStr := params.Get("_offset")
	if offsetStr != "" {
		offset, err = strconv.ParseInt(offsetStr, 10, 64)
		if err != nil {
			return nil, err
		}
		params.Del("_offset")
	}

	for _, obj := range list {
		if cfg.mayRead != nil {
			if cfg.mayRead(obj, r) != nil {
				continue
			}
		}

		matches, err := match(obj, params)
		if err != nil {
			return nil, err
		}
		if !matches {
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
