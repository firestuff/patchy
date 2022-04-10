package api

import "fmt"
import "testing"

import "github.com/go-resty/resty/v2"
import "github.com/stretchr/testify/require"

func TestPUT(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
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

		replaced := &testType{}

		resp, err = c.R().
			SetBody(&testType{
				Text: "bar",
			}).
			SetResult(replaced).
			Put(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "bar", replaced.Text)
		require.Equal(t, created.Id, replaced.Id)

		read := &testType{}

		resp, err = c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "bar", read.Text)
		require.Equal(t, int64(0), read.Num)
		require.Equal(t, created.Id, read.Id)
	})
}

func TestPUTIfMatch(t *testing.T) {
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

		replaced := &testType{}

		resp, err = c.R().
			SetBody(&testType{
				Text: "bar",
			}).
			SetResult(replaced).
			Put(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		etag = resp.Header().Get("ETag")
		require.Equal(t, fmt.Sprintf(`"%s"`, replaced.ETag), etag)
		require.NotEqual(t, created.ETag, replaced.ETag)

		resp, err = c.R().
			SetHeader("If-Match", `"foobar"`).
			SetBody(&testType{
				Text: "zig",
			}).
			SetResult(replaced).
			Put(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.True(t, resp.IsError())
		require.Equal(t, 412, resp.StatusCode())
	})
}
