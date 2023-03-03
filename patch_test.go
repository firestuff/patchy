package patchy_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPATCH(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	created := &testType{}

	resp, err := ta.r().
		SetBody(&testType{
			Text: "foo",
		}).
		SetResult(created).
		Post("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, int64(0), created.Generation)

	updated := &testType{}

	resp, err = ta.r().
		SetBody(&testType{
			Text: "bar",
		}).
		SetResult(updated).
		SetPathParam("id", created.ID).
		Patch("testtype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "bar", updated.Text)
	require.Equal(t, created.ID, updated.ID)
	require.Equal(t, int64(1), updated.Generation)

	read := &testType{}

	resp, err = ta.r().
		SetResult(read).
		SetPathParam("id", created.ID).
		Get("testtype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "bar", read.Text)
	require.Equal(t, created.ID, read.ID)
	require.Equal(t, int64(1), read.Generation)
}

func TestPATCHIfMatch(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	created := &testType{}

	resp, err := ta.r().
		SetBody(&testType{
			Text: "foo",
		}).
		SetResult(created).
		Post("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, int64(0), created.Generation)

	etag := resp.Header().Get("ETag")
	require.Equal(t, fmt.Sprintf(`"%s"`, created.ETag), etag)

	updated := &testType{}

	resp, err = ta.r().
		SetHeader("If-Match", etag).
		SetBody(&testType{
			Text: "bar",
		}).
		SetResult(updated).
		SetPathParam("id", created.ID).
		Patch("testtype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, int64(1), updated.Generation)

	etag = resp.Header().Get("ETag")
	require.Equal(t, fmt.Sprintf(`"%s"`, updated.ETag), etag)
	require.NotEqual(t, created.ETag, updated.ETag)

	resp, err = ta.r().
		SetHeader("If-Match", `"foobar"`).
		SetBody(&testType{
			Text: "zig",
		}).
		SetPathParam("id", created.ID).
		Patch("testtype/{id}")
	require.Nil(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, 400, resp.StatusCode())

	resp, err = ta.r().
		SetHeader("If-Match", `"etag:foobar"`).
		SetBody(&testType{
			Text: "zig",
		}).
		SetPathParam("id", created.ID).
		Patch("testtype/{id}")
	require.Nil(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, 412, resp.StatusCode())

	resp, err = ta.r().
		SetHeader("If-Match", `"generation:1"`).
		SetBody(&testType{
			Text: "zig",
		}).
		SetResult(updated).
		SetPathParam("id", created.ID).
		Patch("testtype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, int64(2), updated.Generation)
	require.Equal(t, "zig", updated.Text)

	resp, err = ta.r().
		SetHeader("If-Match", `"generation:1"`).
		SetBody(&testType{
			Text: "zag",
		}).
		SetPathParam("id", created.ID).
		Patch("testtype/{id}")
	require.Nil(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, 412, resp.StatusCode())
}
