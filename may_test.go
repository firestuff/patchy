//nolint:goerr113
package patchy_test

import (
	"bufio"
	"fmt"
	"net/http"
	"testing"

	"github.com/firestuff/patchy"
	"github.com/stretchr/testify/require"
)

type mayType struct {
	patchy.Metadata
	Text1 string
}

func (mt *mayType) MayRead(r *http.Request) error {
	if r.Header.Get("X-Refuse-Read") != "" {
		return fmt.Errorf("may not read")
	}

	text1 := r.Header.Get("X-Text1")
	if text1 != "" {
		mt.Text1 = text1
	}

	return nil
}

func (*mayType) MayWrite(prev *mayType, r *http.Request) error {
	if r.Header.Get("X-Refuse-Write") != "" {
		return fmt.Errorf("may not write")
	}

	return nil
}

func TestMayWrite(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	patchy.Register[mayType](ta.api)

	created := &mayType{}

	resp, err := ta.r().
		SetBody(&mayType{}).
		SetResult(created).
		Post("maytype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.NotEmpty(t, created.ID)

	resp, err = ta.r().
		SetHeader("X-Refuse-Write", "x").
		SetBody(&mayType{}).
		SetResult(created).
		Post("maytype")
	require.Nil(t, err)
	require.True(t, resp.IsError())

	replaced := &mayType{}

	resp, err = ta.r().
		SetBody(&mayType{}).
		SetResult(replaced).
		SetPathParam("id", created.ID).
		Put("maytype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	resp, err = ta.r().
		SetHeader("X-Refuse-Write", "x").
		SetBody(&mayType{}).
		SetResult(replaced).
		SetPathParam("id", created.ID).
		Put("maytype/{id}")
	require.Nil(t, err)
	require.True(t, resp.IsError())

	updated := &mayType{}

	resp, err = ta.r().
		SetBody(&mayType{}).
		SetResult(updated).
		SetPathParam("id", created.ID).
		Patch("maytype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError(), resp)

	resp, err = ta.r().
		SetHeader("X-Refuse-Write", "x").
		SetBody(&mayType{}).
		SetResult(updated).
		SetPathParam("id", created.ID).
		Patch("maytype/{id}")
	require.Nil(t, err)
	require.True(t, resp.IsError())

	resp, err = ta.r().
		SetHeader("X-Refuse-Write", "x").
		SetPathParam("id", created.ID).
		Delete("maytype/{id}")
	require.Nil(t, err)
	require.True(t, resp.IsError())

	resp, err = ta.r().
		SetPathParam("id", created.ID).
		Delete("maytype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
}

func TestMayRead(t *testing.T) {
	// TODO: list
	// TODO: stream list
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	patchy.Register[mayType](ta.api)

	created := &mayType{}

	resp, err := ta.r().
		SetBody(&mayType{}).
		SetResult(created).
		Post("maytype")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	read := &testType{}

	resp, err = ta.r().
		SetResult(read).
		SetPathParam("id", created.ID).
		Get("maytype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetPathParam("id", created.ID).
		Get("maytype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	resp.RawBody().Close()

	resp, err = ta.r().
		SetHeader("X-Refuse-Read", "x").
		SetResult(read).
		SetPathParam("id", created.ID).
		Get("maytype/{id}")
	require.Nil(t, err)
	require.True(t, resp.IsError())

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("X-Refuse-Read", "x").
		SetHeader("Accept", "text/event-stream").
		SetPathParam("id", created.ID).
		Get("maytype/{id}")
	require.Nil(t, err)
	require.True(t, resp.IsError())
}

func TestMayReadMutate(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	patchy.Register[mayType](ta.api)

	create := &mayType{}

	resp, err := ta.r().
		SetHeader("X-Text1", "1234").
		SetBody(&mayType{Text1: "foo"}).
		SetResult(create).
		Post("maytype")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "1234", create.Text1)

	get := &mayType{}

	resp, err = ta.r().
		SetResult(get).
		SetPathParam("id", create.ID).
		Get("maytype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "foo", get.Text1)

	patch := &mayType{}

	resp, err = ta.r().
		SetHeader("X-Text1", "2345").
		SetBody(&mayType{Text1: "bar"}).
		SetResult(patch).
		SetPathParam("id", create.ID).
		Patch("maytype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "2345", patch.Text1)

	resp, err = ta.r().
		SetResult(get).
		SetPathParam("id", create.ID).
		Get("maytype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "bar", get.Text1)

	resp, err = ta.r().
		SetDoNotParseResponse(true).
		SetHeader("X-Text1", "stream1234").
		SetHeader("Accept", "text/event-stream").
		SetPathParam("id", create.ID).
		Get("maytype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())

	body1 := resp.RawBody()
	defer body1.Close()

	scan1 := bufio.NewScanner(body1)

	stream := &mayType{}

	eventType, err := readEvent(scan1, stream)
	require.Nil(t, err)
	require.Equal(t, "initial", eventType)
	require.Equal(t, "stream1234", stream.Text1)

	put := &mayType{}

	resp, err = ta.r().
		SetHeader("X-Text1", "3456").
		SetBody(&mayType{Text1: "zig"}).
		SetResult(put).
		SetPathParam("id", create.ID).
		Put("maytype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "3456", put.Text1)

	resp, err = ta.r().
		SetResult(get).
		SetPathParam("id", create.ID).
		Get("maytype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "zig", get.Text1)

	resp, err = ta.r().
		SetHeader("X-Text1", "4567").
		SetResult(get).
		SetPathParam("id", create.ID).
		Get("maytype/{id}")
	require.Nil(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "4567", get.Text1)

	eventType, err = readEvent(scan1, stream)
	require.Nil(t, err)
	require.Equal(t, "update", eventType)
	require.Equal(t, "stream1234", stream.Text1)
}
