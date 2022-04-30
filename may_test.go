// nolint:goerr113
package patchy_test

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/firestuff/patchy"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

type flagType struct {
	patchy.Metadata
}

var (
	mayCreateFlag  bool
	mayReplaceFlag bool
	mayUpdateFlag  bool
	mayDeleteFlag  bool
	mayReadFlag    bool
)

var flagMu sync.Mutex

func (*flagType) MayCreate(r *http.Request) error {
	flagMu.Lock()
	defer flagMu.Unlock()

	if !mayCreateFlag {
		return fmt.Errorf("may not create")
	}

	return nil
}

func (*flagType) MayReplace(replace *flagType, r *http.Request) error {
	flagMu.Lock()
	defer flagMu.Unlock()

	if !mayReplaceFlag {
		return fmt.Errorf("may not replace")
	}

	return nil
}

func (*flagType) MayUpdate(patch *flagType, r *http.Request) error {
	flagMu.Lock()
	defer flagMu.Unlock()

	if !mayUpdateFlag {
		return fmt.Errorf("may not update")
	}

	return nil
}

func (*flagType) MayDelete(r *http.Request) error {
	flagMu.Lock()
	defer flagMu.Unlock()

	if !mayDeleteFlag {
		return fmt.Errorf("may not delete")
	}

	return nil
}

func (*flagType) MayRead(r *http.Request) error {
	flagMu.Lock()
	defer flagMu.Unlock()

	if !mayReadFlag {
		return fmt.Errorf("may not read")
	}

	return nil
}

func TestMayCreate(t *testing.T) { // nolint:paralleltest
	withAPI(t, func(t *testing.T, api *patchy.API, baseURL string, c *resty.Client) {
		patchy.Register[flagType](api)

		created := &flagType{}

		flagMu.Lock()
		mayCreateFlag = true
		flagMu.Unlock()

		resp, err := c.R().
			SetBody(&flagType{}).
			SetResult(created).
			Post(fmt.Sprintf("%s/flagtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		require.NotEmpty(t, created.ID)

		flagMu.Lock()
		mayCreateFlag = false
		flagMu.Unlock()

		resp, err = c.R().
			SetBody(&flagType{}).
			SetResult(created).
			Post(fmt.Sprintf("%s/flagtype", baseURL))
		require.Nil(t, err)
		require.True(t, resp.IsError())
	})
}

func TestMayReplace(t *testing.T) { // nolint:paralleltest
	withAPI(t, func(t *testing.T, api *patchy.API, baseURL string, c *resty.Client) {
		patchy.Register[flagType](api)

		created := &flagType{}

		flagMu.Lock()
		mayCreateFlag = true
		flagMu.Unlock()

		resp, err := c.R().
			SetBody(&flagType{}).
			SetResult(created).
			Post(fmt.Sprintf("%s/flagtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		flagMu.Lock()
		mayReplaceFlag = true
		flagMu.Unlock()

		replaced := &flagType{}

		resp, err = c.R().
			SetBody(&flagType{}).
			SetResult(replaced).
			Put(fmt.Sprintf("%s/flagtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		flagMu.Lock()
		mayReplaceFlag = false
		flagMu.Unlock()

		resp, err = c.R().
			SetBody(&flagType{}).
			SetResult(replaced).
			Put(fmt.Sprintf("%s/flagtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.True(t, resp.IsError())
	})
}

func TestMayUpdate(t *testing.T) { // nolint:paralleltest
	withAPI(t, func(t *testing.T, api *patchy.API, baseURL string, c *resty.Client) {
		patchy.Register[flagType](api)

		created := &flagType{}

		flagMu.Lock()
		mayCreateFlag = true
		flagMu.Unlock()

		resp, err := c.R().
			SetBody(&flagType{}).
			SetResult(created).
			Post(fmt.Sprintf("%s/flagtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		flagMu.Lock()
		mayUpdateFlag = true
		flagMu.Unlock()

		updated := &flagType{}

		resp, err = c.R().
			SetBody(&flagType{}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/flagtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		flagMu.Lock()
		mayUpdateFlag = false
		flagMu.Unlock()

		resp, err = c.R().
			SetBody(&flagType{}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/flagtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.True(t, resp.IsError())
	})
}

func TestMayDelete(t *testing.T) { // nolint:paralleltest
	withAPI(t, func(t *testing.T, api *patchy.API, baseURL string, c *resty.Client) {
		patchy.Register[flagType](api)

		created := &flagType{}

		flagMu.Lock()
		mayCreateFlag = true
		flagMu.Unlock()

		resp, err := c.R().
			SetBody(&flagType{}).
			SetResult(created).
			Post(fmt.Sprintf("%s/flagtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		flagMu.Lock()
		mayDeleteFlag = false
		flagMu.Unlock()

		resp, err = c.R().
			Delete(fmt.Sprintf("%s/flagtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.True(t, resp.IsError())

		flagMu.Lock()
		mayDeleteFlag = true
		flagMu.Unlock()

		resp, err = c.R().
			Delete(fmt.Sprintf("%s/flagtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.False(t, resp.IsError())
	})
}

func TestMayRead(t *testing.T) { // nolint:paralleltest
	withAPI(t, func(t *testing.T, api *patchy.API, baseURL string, c *resty.Client) {
		patchy.Register[flagType](api)

		created := &flagType{}

		flagMu.Lock()
		mayCreateFlag = true
		flagMu.Unlock()

		resp, err := c.R().
			SetBody(&flagType{}).
			SetResult(created).
			Post(fmt.Sprintf("%s/flagtype", baseURL))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		read := &testType{}

		flagMu.Lock()
		mayReadFlag = true
		flagMu.Unlock()

		resp, err = c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/flagtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		resp, err = c.R().
			SetDoNotParseResponse(true).
			SetHeader("Accept", "text/event-stream").
			Get(fmt.Sprintf("%s/flagtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		resp.RawBody().Close()

		flagMu.Lock()
		mayReadFlag = false
		flagMu.Unlock()

		resp, err = c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/flagtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.True(t, resp.IsError())

		resp, err = c.R().
			SetDoNotParseResponse(true).
			SetHeader("Accept", "text/event-stream").
			Get(fmt.Sprintf("%s/flagtype/%s", baseURL, created.ID))
		require.Nil(t, err)
		require.True(t, resp.IsError())
	})
}
