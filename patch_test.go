package api

import "fmt"
import "testing"

import "github.com/go-resty/resty/v2"

func TestPATCH(t *testing.T) {
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

		updated := &testType{}

		patch, err := c.R().
			SetBody(&testType{
				Text: "bar",
			}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}
		if patch.IsError() {
			t.Fatal(patch)
		}

		if updated.Text != "bar" {
			t.Fatalf("%s", updated.Text)
		}

		if updated.Id != created.Id {
			t.Fatalf("%s %s", updated.Id, created.Id)
		}

		read := &testType{}

		_, err = c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}

		if read.Text != "bar" {
			t.Fatalf("%s", read.Text)
		}

		if read.Id != created.Id {
			t.Fatalf("%s %s", read.Id, created.Id)
		}
	})
}

func TestPATCHIfMatch(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		created := &testType{}

		resp, err := c.R().
			SetBody(&testType{
				Text: "foo",
			}).
			SetResult(created).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		if err != nil {
			t.Fatal(err)
		}

		etag := resp.Header().Get("ETag")
		if etag != fmt.Sprintf(`"%s"`, created.ETag) {
			t.Fatalf("%s vs %+v", etag, created)
		}

		updated := &testType{}

		resp, err = c.R().
			SetBody(&testType{
				Text: "bar",
			}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}

		etag = resp.Header().Get("ETag")
		if etag != fmt.Sprintf(`"%s"`, updated.ETag) {
			t.Fatalf("%s vs %+v", etag, updated)
		}

		if updated.ETag == created.ETag {
			t.Fatalf("ETag unchanged")
		}

		resp, err = c.R().
			SetHeader("If-Match", `"foobar"`).
			SetBody(&testType{
				Text: "zig",
			}).
			SetResult(updated).
			Patch(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}
		if !resp.IsError() {
			t.Fatal("improper success")
		}
		if resp.StatusCode() != 412 {
			t.Fatal(resp.StatusCode())
		}
	})
}
