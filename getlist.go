package api

import "net/http"

func (api *API) getList(cfg *config, w http.ResponseWriter, r *http.Request) {
	list, err := api.sb.List(cfg.typeName, cfg.factory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	err = writeJsonList(w, ret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
