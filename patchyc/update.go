package patchyc

import (
	"fmt"

	"github.com/firestuff/patchy/api"
	"github.com/firestuff/patchy/metadata"
	"github.com/go-resty/resty/v2"
)

type (
	UpdateOpts = api.UpdateOpts
)

func applyUpdateOpts(opts *UpdateOpts, req *resty.Request) {
	if opts.Prev != nil {
		md := metadata.GetMetadata(opts.Prev)
		req.SetHeader("If-Match", fmt.Sprintf(`"%s"`, md.ETag))
	}
}
