package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

var (
	ErrInvalidIfMatch           = errors.New("invalid If-Match")
	ErrIfMatchMissingQuotes     = fmt.Errorf("missing quotes (%w)", ErrInvalidIfMatch)
	ErrIfMatchUnknownType       = fmt.Errorf("unknown type (%w)", ErrInvalidIfMatch)
	ErrIfMatchInvalidGeneration = fmt.Errorf("invalid generation (%w)", ErrInvalidIfMatch)

	ErrMismatch           = errors.New("If-Match mismatch")
	ErrEtagMismatch       = fmt.Errorf("etag mismatch (%w)", ErrMismatch)
	ErrGenerationMismatch = fmt.Errorf("generation mismatch (%w)", ErrMismatch)
)

func ifMatch(obj any, match string) error {
	// TODO: Support a better intermediate representation for If-Match values, so they can be used by direct clients easily
	if match == "" {
		return nil
	}

	objMD := metadata.GetMetadata(obj)

	if len(match) < 2 || !strings.HasPrefix(match, `"`) || !strings.HasSuffix(match, `"`) {
		return jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", match, ErrIfMatchMissingQuotes)
	}

	val := strings.TrimPrefix(strings.TrimSuffix(match, `"`), `"`)

	switch {
	case strings.HasPrefix(val, "etag:"):
		if val != objMD.ETag {
			return jsrest.Errorf(jsrest.ErrPreconditionFailed, "%s vs %s (%w)", val, objMD.ETag, ErrEtagMismatch)
		}

		return nil

	case strings.HasPrefix(val, "generation:"):
		gen, err := strconv.ParseInt(strings.TrimPrefix(val, "generation:"), 10, 64)
		if err != nil {
			return jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", match, ErrIfMatchInvalidGeneration)
		}

		if gen != objMD.Generation {
			return jsrest.Errorf(jsrest.ErrPreconditionFailed, "%d vs %d (%w)", gen, objMD.Generation, ErrGenerationMismatch)
		}

		return nil

	default:
		return jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", match, ErrIfMatchUnknownType)
	}
}
