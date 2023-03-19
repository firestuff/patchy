package header_test

import (
	"net/http"
	"testing"

	"github.com/firestuff/patchy/header"
	"github.com/stretchr/testify/require"
)

func TestParseAuthorization(t *testing.T) {
	req, err := http.NewRequest("GET", "/foo", nil)
	require.NoError(t, err)

	req.Header.Add("Authorization", "Bearer xyz")

	typ, val, err := header.ParseAuthorization(req)
	require.NoError(t, err)
	require.Equal(t, "Bearer", typ)
	require.Equal(t, "xyz", val)
}

func TestParseBasic(t *testing.T) {
	req, err := http.NewRequest("GET", "/foo", nil)
	require.NoError(t, err)

	req.Header.Add("Authorization", "Basic YWxhZGRpbjpvcGVuc2VzYW1l")

	typ, val, err := header.ParseAuthorization(req)
	require.NoError(t, err)
	require.Equal(t, "Basic", typ)

	user, pass, err := header.ParseBasic(val)
	require.NoError(t, err)
	require.Equal(t, "aladdin", user)
	require.Equal(t, "opensesame", pass)
}
