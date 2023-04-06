package api

import (
	"fmt"
	"net/http"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
	"github.com/vfaronov/httpheader"
)

type UpdateOpts struct {
	IfMatch []httpheader.EntityTag `json:"-"`

	Prev any `json:"prev"`
}

func parseUpdateOpts(r *http.Request) *UpdateOpts {
	return &UpdateOpts{
		IfMatch: httpheader.IfMatch(r.Header),
	}
}

func (opts *UpdateOpts) ifMatch(obj any) error {
	if len(opts.IfMatch) == 0 {
		return nil
	}

	md := metadata.GetMetadata(obj)
	gen := fmt.Sprintf("generation:%d", md.Generation)

	if httpheader.Match(opts.IfMatch, httpheader.EntityTag{Opaque: md.ETag}) ||
		httpheader.Match(opts.IfMatch, httpheader.EntityTag{Opaque: gen}) {
		return nil
	}

	return jsrest.Errorf(jsrest.ErrPreconditionFailed, "If-Match mismatch")
}
