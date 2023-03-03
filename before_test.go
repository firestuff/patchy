package patchy_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/firestuff/patchy"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

type beforeType struct {
	patchy.Metadata
	Text1 string
	Text2 string
}

func (bt *beforeType) BeforeRead(r *http.Request) error {
	bt.Text1 = r.Header.Get("X-Test")

	return nil
}

func TestBeforeRead(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *patchy.API, baseURL string, c *resty.Client) {
		patchy.Register[beforeType](api)

		created := &beforeType{}

		resp, err := c.R().
			SetHeader("X-Test", "1234").
			SetBody(&beforeType{}).
			SetResult(created).
			Post(fmt.Sprintf("%s/beforetype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "1234", created.Text1)

		patched := &beforeType{}

		resp, err = c.R().
			SetHeader("X-Test", "2345").
			SetBody(&beforeType{}).
			SetResult(patched).
			Patch(fmt.Sprintf("%s/beforetype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "2345", patched.Text1)

		// TODO: PUT
		// TODO: GET
		// TODO: list
		// TODO: stream
	})
}
