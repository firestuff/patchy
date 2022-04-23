package api

import "fmt"
import "testing"

import "github.com/go-resty/resty/v2"
import "github.com/stretchr/testify/require"

func TestGETList(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		created1 := &testType{}

		resp, err := c.R().
			SetBody(&testType{
				Text: "foo",
			}).
			SetResult(created1).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		created2 := &testType{}

		resp, err = c.R().
			SetBody(&testType{
				Text: "bar",
			}).
			SetResult(created2).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		list := []testType{}

		resp, err = c.R().
			SetResult(&list).
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 2)
		require.ElementsMatch(t, []string{"foo", "bar"}, []string{list[0].Text, list[1].Text})

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("text", "bar").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 1)
		require.ElementsMatch(t, []string{"bar"}, []string{list[0].Text})

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("_limit", "2").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 2)
		require.ElementsMatch(t, []string{"foo", "bar"}, []string{list[0].Text, list[1].Text})

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("_limit", "1").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 1)
		require.True(t, list[0].Text == "foo" || list[0].Text == "bar")
	})
}
