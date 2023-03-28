package patchyc

import (
	"fmt"

	"github.com/firestuff/patchy/api"
	"github.com/firestuff/patchy/metadata"
	"github.com/go-resty/resty/v2"
)

type (
	GetOpts = api.GetOpts
)

func applyGetOpts(opts *GetOpts, req *resty.Request) {
	if opts.Prev != nil {
		md := metadata.GetMetadata(opts.Prev)
		req.SetHeader("If-None-Match", fmt.Sprintf(`"%s"`, md.ETag))
	}
}
