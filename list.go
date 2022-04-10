package api

import "net/http"

func (api *API) list(cfg *config, r *http.Request) ([]any, error) {
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

		ret = append(ret, obj)
	}

	return ret, nil
}
