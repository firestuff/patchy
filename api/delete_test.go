package api

import "bufio"
import "fmt"
import "testing"

import "github.com/go-resty/resty/v2"

func TestDELETE(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		created := &testType{}

		_, err := c.R().
			SetBody(&testType{
				Text: "foo",
			}).
			SetResult(created).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		if err != nil {
			t.Fatal(err)
		}

		resp, err := c.R().
			SetDoNotParseResponse(true).
			SetHeader("Accept", "text/event-stream").
			Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}
		body := resp.RawBody()
		defer body.Close()

		scan := bufio.NewScanner(body)

		initial := &testType{}
		eventType, err := readEvent(scan, initial)
		if err != nil {
			t.Fatal(err)
		}

		if eventType != "update" {
			t.Error(eventType)
		}

		if initial.Text != "foo" {
			t.Error(initial)
		}

		_, err = c.R().
			Delete(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}

		eventType, err = readEvent(scan, nil)

		if eventType != "delete" {
			t.Error(eventType)
		}

		body.Close()

		read := &testType{}

		resp, err = c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if !resp.IsError() {
			t.Fatal(read)
		}
	})
}
