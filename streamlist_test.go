package api

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
	})
}
