package api

import "fmt"
import "testing"

import "github.com/go-resty/resty/v2"
import "github.com/stretchr/testify/require"

func TestPATCH(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
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

		updated := &testType{}

		resp, err = c.R().
			SetBody(&testType{
				Text: "bar",
			}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "bar", updated.Text)
		require.Equal(t, created.Id, updated.Id)
		require.Equal(t, int64(1), updated.Generation)

		read := &testType{}

		resp, err = c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "bar", read.Text)
		require.Equal(t, created.Id, read.Id)
		require.Equal(t, int64(1), read.Generation)
	})
}

func TestPATCHIfMatch(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
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

		updated := &testType{}

		resp, err = c.R().
			SetHeader("If-Match", etag).
			SetBody(&testType{
				Text: "bar",
			}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, int64(1), updated.Generation)

		etag = resp.Header().Get("ETag")
		require.Equal(t, fmt.Sprintf(`"%s"`, updated.ETag), etag)
		require.NotEqual(t, created.ETag, updated.ETag)

		resp, err = c.R().
			SetHeader("If-Match", `"foobar"`).
			SetBody(&testType{
				Text: "zig",
			}).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.True(t, resp.IsError())
		require.Equal(t, 400, resp.StatusCode())

		resp, err = c.R().
			SetHeader("If-Match", `"etag:foobar"`).
			SetBody(&testType{
				Text: "zig",
			}).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.True(t, resp.IsError())
		require.Equal(t, 412, resp.StatusCode())

		resp, err = c.R().
			SetHeader("If-Match", `"generation:1"`).
			SetBody(&testType{
				Text: "zig",
			}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, int64(2), updated.Generation)
		require.Equal(t, "zig", updated.Text)

		resp, err = c.R().
			SetHeader("If-Match", `"generation:1"`).
			SetBody(&testType{
				Text: "zag",
			}).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.True(t, resp.IsError())
		require.Equal(t, 412, resp.StatusCode())
	})
}
