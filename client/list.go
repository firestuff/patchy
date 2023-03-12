package client

import (
	"fmt"
	"net/url"

	"github.com/firestuff/patchy/api"
	"github.com/go-resty/resty/v2"
)

type (
	ListOpts = api.ListOpts
	Filter   = api.Filter
)

func applyListOpts(opts *ListOpts, req *resty.Request) {
	if opts.Limit != 0 {
		req.SetQueryParam("_limit", fmt.Sprintf("%d", opts.Limit))
	}

	if opts.Offset != 0 {
		req.SetQueryParam("_offset", fmt.Sprintf("%d", opts.Offset))
	}

	if opts.After != "" {
		req.SetQueryParam("_after", opts.After)
	}

	for _, filter := range opts.Filters {
		req.SetQueryParam(fmt.Sprintf("%s[%s]", filter.Path, filter.Op), filter.Value)
	}

	sorts := url.Values{}

	for _, sort := range opts.Sorts {
		sorts.Add("_sort", sort)
	}

	req.SetQueryParamsFromValues(sorts)
}
