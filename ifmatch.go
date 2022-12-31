package patchy

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
)

var (
	ErrInvalidIfMatch           = errors.New("invalid If-Match")
	ErrIfMatchMissingQuotes     = fmt.Errorf("missing quotes: %w", ErrInvalidIfMatch)
	ErrIfMatchUnknownType       = fmt.Errorf("unknown type: %w", ErrInvalidIfMatch)
	ErrIfMatchInvalidGeneration = fmt.Errorf("invalid generation: %w", ErrInvalidIfMatch)

	ErrMismatch           = errors.New("If-Match mismatch")
	ErrEtagMismatch       = fmt.Errorf("etag mismatch: %w", ErrMismatch)
	ErrGenerationMismatch = fmt.Errorf("generation mismatch: %w", ErrMismatch)
)

func ifMatch(obj any, r *http.Request) *jsrest.Error {
	match := r.Header.Get("If-Match")
	if match == "" {
		return nil
	}

	objMD := metadata.GetMetadata(obj)

	if len(match) < 2 || !strings.HasPrefix(match, `"`) || !strings.HasSuffix(match, `"`) {
		e := fmt.Errorf("%s: %w", match, ErrIfMatchMissingQuotes)
		jse := jsrest.FromError(e, jsrest.StatusBadRequest)

		return jse
	}

	val := strings.TrimPrefix(strings.TrimSuffix(match, `"`), `"`)

	switch {
	case strings.HasPrefix(val, "etag:"):
		if val != objMD.ETag {
			e := fmt.Errorf("%s vs %s: %w", val, objMD.ETag, ErrEtagMismatch)
			jse := jsrest.FromError(e, jsrest.StatusPreconditionFailed)

			return jse
		}

		return nil

	case strings.HasPrefix(val, "generation:"):
		gen, err := strconv.ParseInt(strings.TrimPrefix(val, "generation:"), 10, 64)
		if err != nil {
			e := fmt.Errorf("%s: %w", match, ErrIfMatchInvalidGeneration)
			jse := jsrest.FromError(e, jsrest.StatusBadRequest)

			return jse
		}

		if gen != objMD.Generation {
			e := fmt.Errorf("%d vs %d: %w", gen, objMD.Generation, ErrGenerationMismatch)
			jse := jsrest.FromError(e, jsrest.StatusPreconditionFailed)

			return jse
		}

		return nil

	default:
		e := fmt.Errorf("%s: %w", match, ErrIfMatchUnknownType)
		jse := jsrest.FromError(e, jsrest.StatusBadRequest)

		return jse
	}
}
