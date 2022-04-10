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

		initials := []*testType{}

		for i := 0; i < 2; i++ {
			initial := &testType{}
			eventType, err := readEvent(scan, initial)
			require.Nil(t, err)
			require.Equal(t, "initial", eventType)

			initials = append(initials, initial)
		}

		require.ElementsMatch(t, []string{"foo", "bar"}, []string{initials[0].Text, initials[1].Text})
	})
}
