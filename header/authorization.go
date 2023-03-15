package header

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/firestuff/patchy/jsrest"
)

func ParseAuthorization(r *http.Request) (string, string, error) {
	auth := r.Header.Get("Authorization")

	if auth == "" {
		return "", "", jsrest.Errorf(jsrest.ErrBadRequest, "Authorization header missing")
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 {
		return "", "", jsrest.Errorf(jsrest.ErrBadRequest, "Authorization header malformed")
	}

	return parts[0], parts[1], nil
}

func ParseBasic(val string) (string, string, error) {
	raw, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return "", "", jsrest.Errorf(jsrest.ErrBadRequest, "Authorization header Basic data base64 decode failed (%w)", err)
	}

	parts := strings.SplitN(string(raw), ":", 2)
	if len(parts) != 2 {
		return "", "", jsrest.Errorf(jsrest.ErrBadRequest, "Authorization header Basic data malformed")
	}

	return parts[0], parts[1], nil
}
