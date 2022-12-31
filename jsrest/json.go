package jsrest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/firestuff/patchy/metadata"
)

var ErrUnsupportedContentType = errors.New("unsupported Content-Type")

func Read(r *http.Request, obj any) *Error {
	// TODO: Parse semicolon params
	switch r.Header.Get("Content-Type") {
	case "":
		fallthrough
	case "application/json":
		break

	default:
		jse := fmt.Errorf("%s: %w", r.Header.Get("Content-Type"), ErrUnsupportedContentType)
		return FromError(jse, http.StatusUnsupportedMediaType)
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(obj)
	if err != nil {
		jse := fmt.Errorf("failed to decode JSON request body: %w", err)
		return FromError(jse, http.StatusBadRequest)
	}

	return nil
}

func Write(w http.ResponseWriter, obj any) *Error {
	m := metadata.GetMetadata(obj)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, m.ETag))

	enc := json.NewEncoder(w)

	err := enc.Encode(obj)
	if err != nil {
		jse := fmt.Errorf("failed to encode JSON response: %w", err)
		return FromError(jse, StatusInternalServerError)
	}

	return nil
}

func WriteList(w http.ResponseWriter, list []any) *Error {
	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)

	err := enc.Encode(list)
	if err != nil {
		jse := fmt.Errorf("failed to encode JSON response: %w", err)
		return FromError(jse, StatusInternalServerError)
	}

	return nil
}
