package patchyc

import (
	"fmt"

	"github.com/firestuff/patchy/api"
	"github.com/go-resty/resty/v2"
)

type (
	UpdateOpts = api.UpdateOpts
)

func applyUpdateOpts(opts *UpdateOpts, req *resty.Request) {
	if opts.IfMatchETag != "" {
		req.SetHeader("If-Match", fmt.Sprintf(`"%s"`, opts.IfMatchETag))
	}

	if opts.IfMatchGeneration > 0 {
		req.SetHeader("If-Match", fmt.Sprintf(`"generation:%d"`, opts.IfMatchGeneration))
	}
}
