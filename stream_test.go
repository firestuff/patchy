package patchy_test

import (
	"bufio"
	"fmt"
	"testing"

	"github.com/firestuff/patchy"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestStream(t *testing.T) {
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

		resp, err = c.R().
			SetDoNotParseResponse(true).
			SetHeader("Accept", "text/event-stream").
			Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		body := resp.RawBody()
		defer body.Close()

		scan := bufio.NewScanner(body)

		initial := &testType{}
		eventType, err := readEvent(scan, initial)
		require.Nil(t, err)
		require.Equal(t, "initial", eventType)
		require.Equal(t, "foo", initial.Text)

		// Heartbeat (after 5 seconds)
		eventType, err = readEvent(scan, nil)
		require.Nil(t, err)
		require.Equal(t, "heartbeat", eventType)

		updated := &testType{}

		// Round trip PATCH -> SSE
		resp, err = c.R().
			SetBody(&testType{
				Text: "bar",
			}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		eventType, err = readEvent(scan, updated)
		require.Nil(t, err)
		require.Equal(t, "update", eventType)
		require.Equal(t, "bar", updated.Text)
	})
}
