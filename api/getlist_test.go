package api_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGETList(t *testing.T) {
	// TODO: Break up test
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	created1 := &testType{}

	resp, err := ta.r().
		SetBody(&testType{
			Text: "foo",
		}).
		SetResult(created1).
		Post("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	created2 := &testType{}

	resp, err = ta.r().
		SetBody(&testType{
			Text: "bar",
		}).
		SetResult(created2).
		Post("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	created3 := &testType{}

	resp, err = ta.r().
		SetBody(&testType{
			Text: "zig",
		}).
		SetResult(created3).
		Post("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	list := []testType{}

	resp, err = ta.r().
		SetResult(&list).
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 3)
	require.ElementsMatch(t, []string{"foo", "bar", "zig"}, []string{list[0].Text, list[1].Text, list[2].Text})

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("text", "bar").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 1)
	require.ElementsMatch(t, []string{"bar"}, []string{list[0].Text})

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("text[eq]", "bar").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 1)
	require.ElementsMatch(t, []string{"bar"}, []string{list[0].Text})

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("text[junk]", "bar").
		Get("testtype")
	require.Nil(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, 400, resp.StatusCode())

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("text[gt]", "foo").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 1)
	require.ElementsMatch(t, []string{"zig"}, []string{list[0].Text})

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("text[gte]", "foo").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 2)
	require.ElementsMatch(t, []string{"foo", "zig"}, []string{list[0].Text, list[1].Text})

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("text[hp]", "f").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 1)
	require.ElementsMatch(t, []string{"foo"}, []string{list[0].Text})

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("text[in]", "zig,foo,zag").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 2)
	require.ElementsMatch(t, []string{"foo", "zig"}, []string{list[0].Text, list[1].Text})

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("text[lt]", "foo").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 1)
	require.ElementsMatch(t, []string{"bar"}, []string{list[0].Text})

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("text[lte]", "foo").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 2)
	require.ElementsMatch(t, []string{"bar", "foo"}, []string{list[0].Text, list[1].Text})

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("_limit", "1").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 1)
	require.True(t, list[0].Text == "bar" || list[0].Text == "foo" || list[0].Text == "zig")

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("_offset", "0").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 3)
	require.ElementsMatch(t, []string{"bar", "foo", "zig"}, []string{list[0].Text, list[1].Text, list[2].Text})

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("_offset", "1").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 2)

	resp, err = ta.r().
		SetResult(&list).
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 3)
	require.ElementsMatch(t, []string{"bar", "foo", "zig"}, []string{list[0].Text, list[1].Text, list[2].Text})

	list2 := []testType{}

	resp, err = ta.r().
		SetResult(&list2).
		SetQueryParam("_after", list[0].ID).
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list2, 2)
	require.ElementsMatch(t, []string{list[1].Text, list[2].Text}, []string{list2[0].Text, list2[1].Text})

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("_sort", "text").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 3)
	require.Equal(t, []string{"bar", "foo", "zig"}, []string{list[0].Text, list[1].Text, list[2].Text})

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("_sort", "+text").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 3)
	require.Equal(t, []string{"bar", "foo", "zig"}, []string{list[0].Text, list[1].Text, list[2].Text})

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("_sort", "-text").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 3)
	require.Equal(t, []string{"zig", "foo", "bar"}, []string{list[0].Text, list[1].Text, list[2].Text})

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("_sort", "+text").
		SetQueryParam("_offset", "1").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 2)
	require.Equal(t, []string{"foo", "zig"}, []string{list[0].Text, list[1].Text})

	resp, err = ta.r().
		SetResult(&list).
		SetQueryParam("_sort", "text").
		SetQueryParam("_limit", "2").
		Get("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Len(t, list, 2)
	require.ElementsMatch(t, []string{"bar", "foo"}, []string{list[0].Text, list[1].Text})
}
