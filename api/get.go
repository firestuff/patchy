package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/firestuff/patchy/jsrest"
)

type GetOpts struct {
	IfNoneMatchETag       string
	IfNoneMatchGeneration int64

	Prev any
}

var (
	ErrInvalidIfNoneMatch           = errors.New("invalid If-None-Match")
	ErrIfNoneMatchUnknownType       = fmt.Errorf("unknown type (%w)", ErrInvalidIfNoneMatch)
	ErrIfNoneMatchInvalidGeneration = fmt.Errorf("invalid generation (%w)", ErrInvalidIfNoneMatch)
)

func parseGetOpts(r *http.Request) (*GetOpts, error) {
	ret := &GetOpts{}

	ifNoneMatch := r.Header.Get("If-None-Match")

	if ifNoneMatch == "" {
		return ret, nil
	}

	val, err := trimQuotes(ifNoneMatch)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrBadRequest, "trim quotes failed (%w) (%w)", err, ErrInvalidIfNoneMatch)
	}

	switch {
	case strings.HasPrefix(val, "etag:"):
		ret.IfNoneMatchETag = val

	case strings.HasPrefix(val, "generation:"):
		gen, err := strconv.ParseInt(strings.TrimPrefix(val, "generation:"), 10, 64)
		if err != nil {
			return nil, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", ifNoneMatch, ErrIfNoneMatchInvalidGeneration)
		}

		ret.IfNoneMatchGeneration = gen

	default:
		return nil, jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", ifNoneMatch, ErrIfNoneMatchUnknownType)
	}

	return ret, nil
}
