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

		read := &testType{}

		resp, err = c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "bar", read.Text)
		require.Equal(t, created.Id, read.Id)
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

		etag := resp.Header().Get("ETag")
		require.Equal(t, fmt.Sprintf(`"%s"`, created.ETag), etag)

		updated := &testType{}

		resp, err = c.R().
			SetBody(&testType{
				Text: "bar",
			}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		etag = resp.Header().Get("ETag")
		require.Equal(t, fmt.Sprintf(`"%s"`, updated.ETag), etag)
		require.NotEqual(t, created.ETag, updated.ETag)

		resp, err = c.R().
			SetHeader("If-Match", `"foobar"`).
			SetBody(&testType{
				Text: "zig",
			}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.True(t, resp.IsError())
		require.Equal(t, 412, resp.StatusCode())
	})
}
