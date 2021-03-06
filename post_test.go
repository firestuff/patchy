package patchy_test

import (
	"fmt"
	"testing"

	"github.com/firestuff/patchy"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestPOST(t *testing.T) {
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
		require.Equal(t, "foo", created.Text)
		require.NotEmpty(t, created.ID)

		read := &testType{}

		resp, err = c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "foo", read.Text)
		require.Equal(t, created.ID, read.ID)
	})
}
