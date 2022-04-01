package api

import "fmt"
import "net/http"
import "sync"
import "testing"

import "github.com/go-resty/resty/v2"

import "github.com/firestuff/patchy/metadata"

type flagType struct {
	metadata.Metadata
}

var mayCreateFlag bool
var mayReplaceFlag bool
var mayUpdateFlag bool
var mayDeleteFlag bool
var mayReadFlag bool

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
		Register[flagType](api, "flagtype")

		created := &flagType{}

		flagMu.Lock()
		mayCreateFlag = true
		flagMu.Unlock()

		resp, err := c.R().
			SetBody(&flagType{}).
			SetResult(created).
			Post(fmt.Sprintf("%s/flagtype", baseURL))
		if err != nil {
			t.Fatal(err)
		}
		if resp.IsError() {
			t.Fatal(resp.Error())
		}

		if created.Id == "" {
			t.Fatal("missing ID")
		}

		flagMu.Lock()
		mayCreateFlag = false
		flagMu.Unlock()

		resp, err = c.R().
			SetBody(&flagType{}).
			SetResult(created).
			Post(fmt.Sprintf("%s/flagtype", baseURL))
		if err != nil {
			t.Fatal(err)
		}
		if !resp.IsError() {
			t.Fatal("improper success")
		}
	})
}

func TestMayReplace(t *testing.T) {
	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		Register[flagType](api, "flagtype")

		created := &flagType{}

		flagMu.Lock()
		mayCreateFlag = true
		flagMu.Unlock()

		_, err := c.R().
			SetBody(&flagType{}).
			SetResult(created).
			Post(fmt.Sprintf("%s/flagtype", baseURL))
		if err != nil {
			t.Fatal(err)
		}

		flagMu.Lock()
		mayReplaceFlag = true
		flagMu.Unlock()

		replaced := &flagType{}

		resp, err := c.R().
			SetBody(&flagType{}).
			SetResult(replaced).
			Put(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}
		if resp.IsError() {
			t.Fatal(resp)
		}

		flagMu.Lock()
		mayReplaceFlag = false
		flagMu.Unlock()

		resp, err = c.R().
			SetBody(&flagType{}).
			SetResult(replaced).
			Put(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}
		if !resp.IsError() {
			t.Fatal("improper success")
		}
	})
}

func TestMayUpdate(t *testing.T) {
	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		Register[flagType](api, "flagtype")

		created := &flagType{}

		flagMu.Lock()
		mayCreateFlag = true
		flagMu.Unlock()

		_, err := c.R().
			SetBody(&flagType{}).
			SetResult(created).
			Post(fmt.Sprintf("%s/flagtype", baseURL))
		if err != nil {
			t.Fatal(err)
		}

		flagMu.Lock()
		mayUpdateFlag = true
		flagMu.Unlock()

		updated := &flagType{}

		resp, err := c.R().
			SetBody(&flagType{}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}
		if resp.IsError() {
			t.Fatal(resp)
		}

		flagMu.Lock()
		mayUpdateFlag = false
		flagMu.Unlock()

		resp, err = c.R().
			SetBody(&flagType{}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}
		if !resp.IsError() {
			t.Fatal("improper success")
		}
	})
}

func TestMayDelete(t *testing.T) {
	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		Register[flagType](api, "flagtype")

		created := &flagType{}

		flagMu.Lock()
		mayCreateFlag = true
		flagMu.Unlock()

		_, err := c.R().
			SetBody(&flagType{}).
			SetResult(created).
			Post(fmt.Sprintf("%s/flagtype", baseURL))
		if err != nil {
			t.Fatal(err)
		}

		flagMu.Lock()
		mayDeleteFlag = false
		flagMu.Unlock()

		resp, err := c.R().
			Delete(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}
		if !resp.IsError() {
			t.Fatal("improper success")
		}

		flagMu.Lock()
		mayDeleteFlag = true
		flagMu.Unlock()

		resp, err = c.R().
			Delete(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}
		if resp.IsError() {
			t.Fatal(resp)
		}
	})
}

func TestMayRead(t *testing.T) {
	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		Register[flagType](api, "flagtype")

		created := &flagType{}

		flagMu.Lock()
		mayCreateFlag = true
		flagMu.Unlock()

		_, err := c.R().
			SetBody(&flagType{}).
			SetResult(created).
			Post(fmt.Sprintf("%s/flagtype", baseURL))
		if err != nil {
			t.Fatal(err)
		}

		read := &testType{}

		flagMu.Lock()
		mayReadFlag = true
		flagMu.Unlock()

		resp, err := c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}
		if resp.IsError() {
			t.Fatal(resp)
		}

		resp, err = c.R().
			SetDoNotParseResponse(true).
			SetHeader("Accept", "text/event-stream").
			Get(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}
		if resp.IsError() {
			t.Fatal(resp)
		}
		resp.RawBody().Close()

		flagMu.Lock()
		mayReadFlag = false
		flagMu.Unlock()

		resp, err = c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}
		if !resp.IsError() {
			t.Fatal("improper success")
		}

		resp, err = c.R().
			SetDoNotParseResponse(true).
			SetHeader("Accept", "text/event-stream").
			Get(fmt.Sprintf("%s/flagtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}
		if !resp.IsError() {
			t.Fatal("improper success")
		}
	})
}
