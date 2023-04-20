package patchyc

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/gopatchy/metadata"
)

type GetOpts struct {
	Prev any
}

func applyGetOpts(opts *GetOpts, req *resty.Request) {
	if opts.Prev != nil {
		md := metadata.GetMetadata(opts.Prev)
		req.SetHeader("If-None-Match", fmt.Sprintf(`"%s"`, md.ETag))
	}
}
