package patchy_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPOST(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	created := &testType{}

	resp, err := ta.r().
		SetBody(&testType{
			Text: "foo",
		}).
		SetResult(created).
		Post("testtype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "foo", created.Text)
	require.NotEmpty(t, created.ID)

	read := &testType{}

	resp, err = ta.r().
		SetResult(read).
		SetPathParam("id", created.ID).
		Get("testtype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "foo", read.Text)
	require.Equal(t, created.ID, read.ID)
}
