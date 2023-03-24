package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

type UpdateOpts struct {
	IfMatchETag       string
	IfMatchGeneration int64
}

var (
	ErrInvalidIfMatch           = errors.New("invalid If-Match")
	ErrIfMatchUnknownType       = fmt.Errorf("unknown type (%w)", ErrInvalidIfMatch)
	ErrIfMatchInvalidGeneration = fmt.Errorf("invalid generation (%w)", ErrInvalidIfMatch)

	ErrMismatch           = errors.New("If-Match mismatch")
	ErrEtagMismatch       = fmt.Errorf("etag mismatch (%w)", ErrMismatch)
	ErrGenerationMismatch = fmt.Errorf("generation mismatch (%w)", ErrMismatch)
)

func parseUpdateOpts(r *http.Request) (*UpdateOpts, error) {
	ret := &UpdateOpts{}

	ifMatch := r.Header.Get("If-Match")

	if ifMatch == "" {
		return ret, nil
	}

	val, err := trimQuotes(ifMatch)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrBadRequest, "trim quotes failed (%w) (%w)", err, ErrInvalidIfMatch)
	}

	switch {
	case strings.HasPrefix(val, "etag:"):
		ret.IfMatchETag = val

	case strings.HasPrefix(val, "generation:"):
		gen, err := strconv.ParseInt(strings.TrimPrefix(val, "generation:"), 10, 64)
		if err != nil {
			return nil, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", ifMatch, ErrIfMatchInvalidGeneration)
		}

		ret.IfMatchGeneration = gen

	default:
		return nil, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", ifMatch, ErrIfMatchUnknownType)
	}

	return ret, nil
}

func ifMatch(obj any, opts *UpdateOpts) error {
	objMD := metadata.GetMetadata(obj)

	if opts.IfMatchETag != "" && opts.IfMatchETag != objMD.ETag {
		return jsrest.Errorf(jsrest.ErrPreconditionFailed, "%s vs %s (%w)", opts.IfMatchETag, objMD.ETag, ErrEtagMismatch)
	}

	if opts.IfMatchGeneration > 0 && opts.IfMatchGeneration != objMD.Generation {
		return jsrest.Errorf(jsrest.ErrPreconditionFailed, "%d vs %d (%w)", opts.IfMatchGeneration, objMD.Generation, ErrGenerationMismatch)
	}

	return nil
}
