package patchy_test

import (
	"testing"

	"github.com/firestuff/patchy"
	"github.com/go-resty/resty/v2"
)

type Outer struct {
	patchy.Metadata
	Text  string
	Inner []string
}

type Inner struct {
	patchy.Metadata
	Text string
}

func TestNested(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *patchy.API, baseURL string, c *resty.Client) {
		patchy.Register[Outer](api)
		patchy.Register[Inner](api)
	})
}
