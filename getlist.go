package api

import "net/http"

func (api *API) getList(cfg *config, w http.ResponseWriter, r *http.Request) {
	list, err := api.list(cfg, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = writeJsonList(w, list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
