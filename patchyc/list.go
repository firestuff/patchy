package patchyc

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"reflect"

	"github.com/go-resty/resty/v2"
	"github.com/gopatchy/jsrest"
	"github.com/gopatchy/metadata"
)

type ListOpts struct {
	Stream  string
	Limit   int64
	Offset  int64
	After   string
	Sorts   []string
	Filters []*Filter

	Prev any
}

type Filter struct {
	Path  string
	Op    string
	Value string
}

func applyListOpts(opts *ListOpts, req *resty.Request) error {
	if opts.Prev != nil {
		etag, err := hashList(opts.Prev)
		if err != nil {
			return err
		}

		req.SetHeader("If-None-Match", fmt.Sprintf(`"%s"`, etag))
	}

	if opts.Stream != "" {
		req.SetQueryParam("_stream", opts.Stream)
	}

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

	return nil
}

func hashList(list any) (string, error) {
	hash := sha256.New()

	v := reflect.ValueOf(list)

	for i := 0; i < v.Len(); i++ {
		iter := v.Index(i)

		md := metadata.GetMetadata(iter.Interface())

		_, err := hash.Write([]byte(md.ETag + "\n"))
		if err != nil {
			return "", jsrest.Errorf(jsrest.ErrInternalServerError, "hash write failed (%w)", err)
		}
	}

	return fmt.Sprintf("etag:%x", hash.Sum(nil)), nil
}
