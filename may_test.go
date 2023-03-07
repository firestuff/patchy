//nolint:goerr113
package patchy_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/firestuff/patchy"
	"github.com/stretchr/testify/require"
)

type mayType struct {
	patchy.Metadata
}

func (*mayType) MayRead(r *http.Request) error {
	if r.Header.Get("X-Refuse-Read") != "" {
		return fmt.Errorf("may not read")
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
