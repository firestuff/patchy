package patchy_test

import (
	"fmt"
	"testing"

	"github.com/firestuff/patchy"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestGETList(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *patchy.API, baseURL string, c *resty.Client) {
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

		created3 := &testType{}

		resp, err = c.R().
			SetBody(&testType{
				Text: "zig",
			}).
			SetResult(created3).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		list := []testType{}

		resp, err = c.R().
			SetResult(&list).
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 3)
		require.ElementsMatch(t, []string{"foo", "bar", "zig"}, []string{list[0].Text, list[1].Text, list[2].Text})

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
			SetQueryParam("text[eq]", "bar").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 1)
		require.ElementsMatch(t, []string{"bar"}, []string{list[0].Text})

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("text[junk]", "bar").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.True(t, resp.IsError())

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("text[gt]", "foo").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 1)
		require.ElementsMatch(t, []string{"zig"}, []string{list[0].Text})

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("text[gte]", "foo").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 2)
		require.ElementsMatch(t, []string{"foo", "zig"}, []string{list[0].Text, list[1].Text})

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("text[in]", "zig,foo,zag").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 2)
		require.ElementsMatch(t, []string{"foo", "zig"}, []string{list[0].Text, list[1].Text})

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("text[lt]", "foo").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 1)
		require.ElementsMatch(t, []string{"bar"}, []string{list[0].Text})

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("text[lte]", "foo").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 2)
		require.ElementsMatch(t, []string{"bar", "foo"}, []string{list[0].Text, list[1].Text})

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("_limit", "1").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 1)
		require.True(t, list[0].Text == "bar" || list[0].Text == "foo" || list[0].Text == "zig")

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("_offset", "0").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 3)
		require.ElementsMatch(t, []string{"bar", "foo", "zig"}, []string{list[0].Text, list[1].Text, list[2].Text})

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("_offset", "1").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 2)

		resp, err = c.R().
			SetResult(&list).
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 3)
		require.ElementsMatch(t, []string{"bar", "foo", "zig"}, []string{list[0].Text, list[1].Text, list[2].Text})

		list2 := []testType{}

		resp, err = c.R().
			SetResult(&list2).
			SetQueryParam("_after", list[0].ID).
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list2, 2)
		require.ElementsMatch(t, []string{list[1].Text, list[2].Text}, []string{list2[0].Text, list2[1].Text})

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("_sort", "text").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 3)
		require.Equal(t, []string{"bar", "foo", "zig"}, []string{list[0].Text, list[1].Text, list[2].Text})

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("_sort", "+text").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 3)
		require.Equal(t, []string{"bar", "foo", "zig"}, []string{list[0].Text, list[1].Text, list[2].Text})

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("_sort", "-text").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 3)
		require.Equal(t, []string{"zig", "foo", "bar"}, []string{list[0].Text, list[1].Text, list[2].Text})

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("_sort", "+text").
			SetQueryParam("_offset", "1").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 2)
		require.Equal(t, []string{"foo", "zig"}, []string{list[0].Text, list[1].Text})

		resp, err = c.R().
			SetResult(&list).
			SetQueryParam("_sort", "text").
			SetQueryParam("_limit", "2").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 2)
		require.ElementsMatch(t, []string{"bar", "foo"}, []string{list[0].Text, list[1].Text})
	})
}
