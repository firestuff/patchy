package patchyc

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/gopatchy/metadata"
)

type UpdateOpts struct {
	Prev any
}

func applyUpdateOpts(opts *UpdateOpts, req *resty.Request) {
	if opts.Prev != nil {
		md := metadata.GetMetadata(opts.Prev)
		req.SetHeader("If-Match", fmt.Sprintf(`"%s"`, md.ETag))
	}
}
