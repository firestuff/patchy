package patchy_test

import (
	"testing"

	"github.com/firestuff/patchy"
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

func TestNestedInnerFirst(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	patchy.Register[Inner](ta.api)
	patchy.Register[Outer](ta.api)
}

func TestNestedOuterFirst(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	patchy.Register[Outer](ta.api)
	patchy.Register[Inner](ta.api)
}
