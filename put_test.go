package patchy_test

import (
	"fmt"
	"testing"

	"github.com/firestuff/patchy"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestPUT(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *patchy.API, baseURL string, c *resty.Client) {
		created := &testType{}

		resp, err := c.R().
			SetBody(&testType{
				Text: "foo",
				Num:  1,
			}).
			SetResult(created).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, int64(0), created.Generation)

		replaced := &testType{}

		resp, err = c.R().
			SetBody(&testType{
				Text: "bar",
			}).
			SetResult(replaced).
			Put(fmt.Sprintf("%s/testtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "bar", replaced.Text)
		require.Equal(t, created.ID, replaced.ID)
		require.Equal(t, int64(1), replaced.Generation)

		read := &testType{}

		resp, err = c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "bar", read.Text)
		require.Equal(t, int64(0), read.Num)
		require.Equal(t, created.ID, read.ID)
		require.Equal(t, int64(1), read.Generation)
	})
}

func TestPUTIfMatch(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *patchy.API, baseURL string, c *resty.Client) {
		created := &testType{}

		resp, err := c.R().
			SetBody(&testType{
				Text: "foo",
			}).
			SetResult(created).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, int64(0), created.Generation)

		etag := resp.Header().Get("ETag")
		require.Equal(t, fmt.Sprintf(`"%s"`, created.ETag), etag)

		replaced := &testType{}

		resp, err = c.R().
			SetHeader("If-Match", etag).
			SetBody(&testType{
				Text: "bar",
			}).
			SetResult(replaced).
			Put(fmt.Sprintf("%s/testtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, int64(1), replaced.Generation)

		etag = resp.Header().Get("ETag")
		require.Equal(t, fmt.Sprintf(`"%s"`, replaced.ETag), etag)
		require.NotEqual(t, created.ETag, replaced.ETag)

		resp, err = c.R().
			SetHeader("If-Match", `"foobar"`).
			SetBody(&testType{
				Text: "zig",
			}).
			Put(fmt.Sprintf("%s/testtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.True(t, resp.IsError())
		require.Equal(t, 400, resp.StatusCode())

		resp, err = c.R().
			SetHeader("If-Match", `"etag:foobar"`).
			SetBody(&testType{
				Text: "zig",
			}).
			Put(fmt.Sprintf("%s/testtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.True(t, resp.IsError())
		require.Equal(t, 412, resp.StatusCode())

		resp, err = c.R().
			SetHeader("If-Match", `"generation:1"`).
			SetBody(&testType{
				Text: "zig",
			}).
			SetResult(replaced).
			Put(fmt.Sprintf("%s/testtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, int64(2), replaced.Generation)
		require.Equal(t, "zig", replaced.Text)

		resp, err = c.R().
			SetHeader("If-Match", `"generation:1"`).
			SetBody(&testType{
				Text: "zag",
			}).
			Put(fmt.Sprintf("%s/testtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.True(t, resp.IsError())
		require.Equal(t, 412, resp.StatusCode())
	})
}
