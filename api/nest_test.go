package api_test

import (
	"testing"

	"github.com/firestuff/patchy/api"
)

type Outer struct {
	api.Metadata
	Text  string
	Inner []string
}

type Inner struct {
	api.Metadata
	Text string
}

func TestNestedInnerFirst(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	api.Register[Inner](ta.api)
	api.Register[Outer](ta.api)
}

func TestNestedOuterFirst(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	api.Register[Outer](ta.api)
	api.Register[Inner](ta.api)
}
