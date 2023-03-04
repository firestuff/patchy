package patchy_test

import (
	"bufio"
	"net/http"
	"testing"

	"github.com/firestuff/patchy"
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
	// TODO: list
	// TODO: stream one
	// TODO: stream list
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	patchy.Register[beforeType](ta.api)

	create := &beforeType{}

	resp, err := ta.r().
		SetHeader("X-Test", "1234").
		SetBody(&beforeType{}).
		SetResult(create).
		Post("beforetype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "1234", create.Text1)

	patch := &beforeType{}

	resp, err = ta.r().
		SetHeader("X-Test", "2345").
		SetBody(&beforeType{}).
		SetResult(patch).
		SetPathParam("id", create.ID).
		Patch("beforetype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "2345", patch.Text1)

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("X-Test", "stream1234").
		SetHeader("Accept", "text/event-stream").
		SetPathParam("id", create.ID).
		Get("beforetype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	body1 := resp.RawBody()
	defer body1.Close()

	scan1 := bufio.NewScanner(body1)

	initial := &beforeType{}
	eventType, err := readEvent(scan1, initial)
	require.Nil(t, err)
	require.Equal(t, "initial", eventType)
	require.Equal(t, "stream1234", initial.Text1)

	put := &beforeType{}

	resp, err = ta.r().
		SetHeader("X-Test", "3456").
		SetBody(&beforeType{}).
		SetResult(put).
		SetPathParam("id", create.ID).
		Put("beforetype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "3456", put.Text1)

	get := &beforeType{}

	resp, err = ta.r().
		SetHeader("X-Test", "4567").
		SetResult(get).
		SetPathParam("id", create.ID).
		Get("beforetype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "4567", get.Text1)
}
