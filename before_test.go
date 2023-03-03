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

		create := &beforeType{}

		resp, err := c.R().
			SetHeader("X-Test", "1234").
			SetBody(&beforeType{}).
			SetResult(create).
			Post(fmt.Sprintf("%s/beforetype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "1234", create.Text1)

		patch := &beforeType{}

		resp, err = c.R().
			SetHeader("X-Test", "2345").
			SetBody(&beforeType{}).
			SetResult(patch).
			Patch(fmt.Sprintf("%s/beforetype/%s", baseURL, create.ID))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "2345", patch.Text1)

		put := &beforeType{}

		resp, err = c.R().
			SetHeader("X-Test", "3456").
			SetBody(&beforeType{}).
			SetResult(put).
			Put(fmt.Sprintf("%s/beforetype/%s", baseURL, create.ID))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "3456", put.Text1)

		get := &beforeType{}

		resp, err = c.R().
			SetHeader("X-Test", "4567").
			SetResult(get).
			Get(fmt.Sprintf("%s/beforetype/%s", baseURL, create.ID))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Equal(t, "4567", get.Text1)

		list := []*beforeType{}

		resp, err = c.R().
			SetHeader("X-Test", "5678").
			SetResult(&list).
			Get(fmt.Sprintf("%s/beforetype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.Len(t, list, 1)
		require.Equal(t, "5678", list[0].Text1)

		// TODO: stream
	})
}
