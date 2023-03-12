package client

import (
	"fmt"
	"net/url"

	"github.com/go-resty/resty/v2"
)

type Matcher struct {
	Path  string
	Op    string
	Value string
}

type ListOptions struct {
	Matchers []*Matcher
	Sort     []string
}

func (lo *ListOptions) Apply(req *resty.Request) {
	for _, matcher := range lo.Matchers {
		req.SetQueryParam(fmt.Sprintf("%s[%s]", matcher.Path, matcher.Op), matcher.Value)
	}

	sorts := url.Values{}

	for _, sort := range lo.Sort {
		sorts.Add("_sort", sort)
	}

	req.SetQueryParamsFromValues(sorts)
}
