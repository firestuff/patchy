package patchy

import "bufio"
import "fmt"
import "testing"

import "github.com/go-resty/resty/v2"
import "github.com/stretchr/testify/require"

func TestStreamList(t *testing.T) {
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

		resp, err = c.R().
			SetDoNotParseResponse(true).
			SetHeader("Accept", "text/event-stream").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		body := resp.RawBody()
		defer body.Close()

		scan := bufio.NewScanner(body)

		list := []*testType{}

		eventType, err := readEvent(scan, &list)
		require.Nil(t, err)
		require.Equal(t, "list", eventType)

		require.Len(t, list, 2)
		require.ElementsMatch(t, []string{"foo", "bar"}, []string{list[0].Text, list[1].Text})

		// Heartbeat (after 5 seconds)
		eventType, err = readEvent(scan, nil)
		require.Equal(t, "heartbeat", eventType)

		created3 := &testType{}

		resp, err = c.R().
			SetBody(&testType{
				Text: "zig",
			}).
			SetResult(created3).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		eventType, err = readEvent(scan, &list)
		require.Nil(t, err)
		require.Equal(t, "list", eventType)

		require.Len(t, list, 3)
		require.ElementsMatch(t, []string{"foo", "bar", "zig"}, []string{list[0].Text, list[1].Text, list[2].Text})

		resp, err = c.R().
			SetBody(&testType{
				Text: "zag",
			}).
			SetResult(created3).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created3.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		eventType, err = readEvent(scan, &list)
		require.Nil(t, err)
		require.Equal(t, "list", eventType)

		require.Len(t, list, 3)
		require.ElementsMatch(t, []string{"foo", "bar", "zag"}, []string{list[0].Text, list[1].Text, list[2].Text})

		resp, err = c.R().
			Delete(fmt.Sprintf("%s/testtype/%s", baseURL, created3.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		eventType, err = readEvent(scan, &list)
		require.Nil(t, err)
		require.Equal(t, "list", eventType)

		require.Len(t, list, 2)
		require.ElementsMatch(t, []string{"foo", "bar"}, []string{list[0].Text, list[1].Text})

		resp, err = c.R().
			SetDoNotParseResponse(true).
			SetHeader("Accept", "text/event-stream").
			SetQueryParam("_limit", "1").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		body2 := resp.RawBody()
		defer body2.Close()

		scan2 := bufio.NewScanner(body2)

		eventType, err = readEvent(scan2, &list)
		require.Nil(t, err)
		require.Equal(t, "list", eventType)

		require.Len(t, list, 1)
		require.True(t, list[0].Text == "foo" || list[0].Text == "bar")
	})
}

func TestStreamListDiff(t *testing.T) {
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

		resp, err = c.R().
			SetDoNotParseResponse(true).
			SetHeader("Accept", "text/event-stream").
			SetQueryParam("_stream", "diff").
			Get(fmt.Sprintf("%s/testtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		body := resp.RawBody()
		defer body.Close()

		scan := bufio.NewScanner(body)

		obj1 := testType{}

		eventType, err := readEvent(scan, &obj1)
		require.Nil(t, err)
		require.Equal(t, "add", eventType)

		obj2 := testType{}

		eventType, err = readEvent(scan, &obj2)
		require.Nil(t, err)
		require.Equal(t, "add", eventType)
		require.ElementsMatch(t, []string{"foo", "bar"}, []string{obj1.Text, obj2.Text})

		resp, err = c.R().
			SetBody(&testType{
				Text: "zig",
			}).
			SetResult(created2).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created2.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		eventType, err = readEvent(scan, &obj1)
		require.Nil(t, err)
		require.Equal(t, "update", eventType)
		require.Equal(t, created2.Id, obj1.Id)
		require.Equal(t, "zig", obj1.Text)

		resp, err = c.R().
			Delete(fmt.Sprintf("%s/testtype/%s", baseURL, created1.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		eventType, err = readEvent(scan, &obj1)
		require.Nil(t, err)
		require.Equal(t, "remove", eventType)
		require.Equal(t, created1.Id, obj1.Id)
		require.Equal(t, "foo", obj1.Text)
	})
}
