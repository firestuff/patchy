package api_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDirect(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	create, err := ta.api.Create("testtype", &testType{Text: "foo"})
	require.Nil(t, err)
	require.Equal(t, "foo", create.(*testType).Text)

	get, err := ta.api.Get("testtype", create.(*testType).ID)
	require.Nil(t, err)
	require.Equal(t, "foo", get.(*testType).Text)
}
