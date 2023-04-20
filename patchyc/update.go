package patchyc

import (
	"fmt"

	"github.com/firestuff/patchy/api"
	"github.com/go-resty/resty/v2"
	"github.com/gopatchy/metadata"
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
