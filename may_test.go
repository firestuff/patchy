//nolint:goerr113
package patchy_test

import (
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/firestuff/patchy"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

type mayType struct {
	patchy.Metadata
}

var (
	mayCreateFlag  bool
	mayReplaceFlag bool
	mayUpdateFlag  bool
	mayDeleteFlag  bool
	mayReadFlag    bool
)

var mayMu sync.Mutex

func (*mayType) MayCreate(r *http.Request) error {
	mayMu.Lock()
	defer mayMu.Unlock()

	if !mayCreateFlag {
		return errors.New("may not create")
	}

	return nil
}

func (*mayType) MayReplace(replace *mayType, r *http.Request) error {
	mayMu.Lock()
	defer mayMu.Unlock()

	if !mayReplaceFlag {
		return errors.New("may not replace")
	}

	return nil
}

func (*mayType) MayUpdate(patch *mayType, r *http.Request) error {
	mayMu.Lock()
	defer mayMu.Unlock()

	if !mayUpdateFlag {
		return errors.New("may not update")
	}

	return nil
}

func (*mayType) MayDelete(r *http.Request) error {
	mayMu.Lock()
	defer mayMu.Unlock()

	if !mayDeleteFlag {
		return errors.New("may not delete")
	}

	return nil
}

func (*mayType) MayRead(r *http.Request) error {
	mayMu.Lock()
	defer mayMu.Unlock()

	if !mayReadFlag {
		return errors.New("may not read")
	}

	return nil
}

func TestMayCreate(t *testing.T) { //nolint:paralleltest
	withAPI(t, func(t *testing.T, api *patchy.API, c *resty.Client) {
		patchy.Register[mayType](api)

		created := &mayType{}

		mayMu.Lock()
		mayCreateFlag = true
		mayMu.Unlock()

		resp, err := c.R().
			SetBody(&mayType{}).
			SetResult(created).
			Post("maytype")
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.NotEmpty(t, created.ID)

		mayMu.Lock()
		mayCreateFlag = false
		mayMu.Unlock()

		resp, err = c.R().
			SetBody(&mayType{}).
			SetResult(created).
			Post("maytype")
		require.Nil(t, err)
		require.True(t, resp.IsError())
	})
}

func TestMayReplace(t *testing.T) { //nolint:paralleltest
	withAPI(t, func(t *testing.T, api *patchy.API, c *resty.Client) {
		patchy.Register[mayType](api)

		created := &mayType{}

		mayMu.Lock()
		mayCreateFlag = true
		mayMu.Unlock()

		resp, err := c.R().
			SetBody(&mayType{}).
			SetResult(created).
			Post("maytype")
		require.Nil(t, err)
		require.False(t, resp.IsError())

		mayMu.Lock()
		mayReplaceFlag = true
		mayMu.Unlock()

		replaced := &mayType{}

		resp, err = c.R().
			SetBody(&mayType{}).
			SetResult(replaced).
			SetPathParam("id", created.ID).
			Put("maytype/{id}")
		require.Nil(t, err)
		require.False(t, resp.IsError())

		mayMu.Lock()
		mayReplaceFlag = false
		mayMu.Unlock()

		resp, err = c.R().
			SetBody(&mayType{}).
			SetResult(replaced).
			SetPathParam("id", created.ID).
			Put("maytype/{id}")
		require.Nil(t, err)
		require.True(t, resp.IsError())
	})
}

func TestMayUpdate(t *testing.T) { //nolint:paralleltest
	withAPI(t, func(t *testing.T, api *patchy.API, c *resty.Client) {
		patchy.Register[mayType](api)

		created := &mayType{}

		mayMu.Lock()
		mayCreateFlag = true
		mayMu.Unlock()

		resp, err := c.R().
			SetBody(&mayType{}).
			SetResult(created).
			Post("maytype")
		require.Nil(t, err)
		require.False(t, resp.IsError())

		mayMu.Lock()
		mayUpdateFlag = true
		mayMu.Unlock()

		updated := &mayType{}

		resp, err = c.R().
			SetBody(&mayType{}).
			SetResult(updated).
			SetPathParam("id", created.ID).
			Patch("maytype/{id}")
		require.Nil(t, err)
		require.False(t, resp.IsError())

		mayMu.Lock()
		mayUpdateFlag = false
		mayMu.Unlock()

		resp, err = c.R().
			SetBody(&mayType{}).
			SetResult(updated).
			SetPathParam("id", created.ID).
			Patch("maytype/{id}")
		require.Nil(t, err)
		require.True(t, resp.IsError())
	})
}

func TestMayDelete(t *testing.T) { //nolint:paralleltest
	withAPI(t, func(t *testing.T, api *patchy.API, c *resty.Client) {
		patchy.Register[mayType](api)

		created := &mayType{}

		mayMu.Lock()
		mayCreateFlag = true
		mayMu.Unlock()

		resp, err := c.R().
			SetBody(&mayType{}).
			SetResult(created).
			Post("maytype")
		require.Nil(t, err)
		require.False(t, resp.IsError())

		mayMu.Lock()
		mayDeleteFlag = false
		mayMu.Unlock()

		resp, err = c.R().
			SetPathParam("id", created.ID).
			Delete("maytype/{id}")
		require.Nil(t, err)
		require.True(t, resp.IsError())

		mayMu.Lock()
		mayDeleteFlag = true
		mayMu.Unlock()

		resp, err = c.R().
			SetPathParam("id", created.ID).
			Delete("maytype/{id}")
		require.Nil(t, err)
		require.False(t, resp.IsError())
	})
}

func TestMayRead(t *testing.T) { //nolint:paralleltest
	withAPI(t, func(t *testing.T, api *patchy.API, c *resty.Client) {
		patchy.Register[mayType](api)

		created := &mayType{}

		mayMu.Lock()
		mayCreateFlag = true
		mayMu.Unlock()

		resp, err := c.R().
			SetBody(&mayType{}).
			SetResult(created).
			Post("maytype")
		require.Nil(t, err)
		require.False(t, resp.IsError())

		read := &testType{}

		mayMu.Lock()
		mayReadFlag = true
		mayMu.Unlock()

		resp, err = c.R().
			SetResult(read).
			SetPathParam("id", created.ID).
			Get("maytype/{id}")
		require.Nil(t, err)
		require.False(t, resp.IsError())

		resp, err = c.R().
			SetDoNotParseResponse(true).
			SetHeader("Accept", "text/event-stream").
			SetPathParam("id", created.ID).
			Get("maytype/{id}")
		require.Nil(t, err)
		require.False(t, resp.IsError())
		resp.RawBody().Close()

		mayMu.Lock()
		mayReadFlag = false
		mayMu.Unlock()

		resp, err = c.R().
			SetResult(read).
			SetPathParam("id", created.ID).
			Get("maytype/{id}")
		require.Nil(t, err)
		require.True(t, resp.IsError())

		resp, err = c.R().
			SetDoNotParseResponse(true).
			SetHeader("Accept", "text/event-stream").
			SetPathParam("id", created.ID).
			Get("maytype/{id}")
		require.Nil(t, err)
		require.True(t, resp.IsError())
	})
}
