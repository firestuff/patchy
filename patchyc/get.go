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
		opts.IfNoneMatchETag = md.ETag
	}

	if opts.IfNoneMatchETag != "" {
		req.SetHeader("If-None-Match", fmt.Sprintf(`"%s"`, opts.IfNoneMatchETag))
	} else if opts.IfNoneMatchGeneration > 0 {
		req.SetHeader("If-None-Match", fmt.Sprintf(`"generation:%d"`, opts.IfNoneMatchGeneration))
	}
}
