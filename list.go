package api

import "net/http"
import "net/url"

import "github.com/firestuff/patchy/path"

func (api *API) list(cfg *config, r *http.Request, params url.Values) ([]any, error) {
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

		matches, err := match(obj, params)
		if err != nil {
			return nil, err
		}
		if !matches {
			continue
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
