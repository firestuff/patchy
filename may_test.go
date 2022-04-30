package patchy

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

type flagType struct {
	Metadata
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

	if mayCreateFlag {
		return nil
	} else {
		return fmt.Errorf("may not create")
	}
}

func (*flagType) MayReplace(replace *flagType, r *http.Request) error {
	flagMu.Lock()
	defer flagMu.Unlock()

	if mayReplaceFlag {
		return nil
	} else {
		return fmt.Errorf("may not replace")
	}
}

func (*flagType) MayUpdate(patch *flagType, r *http.Request) error {
	flagMu.Lock()
	defer flagMu.Unlock()

	if mayUpdateFlag {
		return nil
	} else {
		return fmt.Errorf("may not update")
	}
}

func (*flagType) MayDelete(r *http.Request) error {
	flagMu.Lock()
	defer flagMu.Unlock()

	if mayDeleteFlag {
		return nil
	} else {
		return fmt.Errorf("may not delete")
	}
}

func (*flagType) MayRead(r *http.Request) error {
	flagMu.Lock()
	defer flagMu.Unlock()

	if mayReadFlag {
		return nil
	} else {
		return fmt.Errorf("may not read")
	}
}

func TestMayCreate(t *testing.T) {
	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		Register[flagType](api)

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
		require.NotEmpty(t, created.Id)

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

func TestMayReplace(t *testing.T) {
	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		Register[flagType](api)

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
			Put(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		flagMu.Lock()
		mayReplaceFlag = false
		flagMu.Unlock()

		resp, err = c.R().
			SetBody(&flagType{}).
			SetResult(replaced).
			Put(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.True(t, resp.IsError())
	})
}

func TestMayUpdate(t *testing.T) {
	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		Register[flagType](api)

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
			Patch(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		flagMu.Lock()
		mayUpdateFlag = false
		flagMu.Unlock()

		resp, err = c.R().
			SetBody(&flagType{}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.True(t, resp.IsError())
	})
}

func TestMayDelete(t *testing.T) {
	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		Register[flagType](api)

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
			Delete(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.True(t, resp.IsError())

		flagMu.Lock()
		mayDeleteFlag = true
		flagMu.Unlock()

		resp, err = c.R().
			Delete(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())
	})
}

func TestMayRead(t *testing.T) {
	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		Register[flagType](api)

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
			Get(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())

		resp, err = c.R().
			SetDoNotParseResponse(true).
			SetHeader("Accept", "text/event-stream").
			Get(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.False(t, resp.IsError())
		resp.RawBody().Close()

		flagMu.Lock()
		mayReadFlag = false
		flagMu.Unlock()

		resp, err = c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.True(t, resp.IsError())

		resp, err = c.R().
			SetDoNotParseResponse(true).
			SetHeader("Accept", "text/event-stream").
			Get(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		require.Nil(t, err)
		require.True(t, resp.IsError())
	})
}
