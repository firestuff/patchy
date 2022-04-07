package api

import "fmt"
import "testing"

import "github.com/go-resty/resty/v2"

func TestPUT(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		created := &testType{}

		_, err := c.R().
			SetBody(&testType{
				Text: "foo",
				Num:  1,
			}).
			SetResult(created).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		if err != nil {
			t.Fatal(err)
		}

		replaced := &testType{}

		put, err := c.R().
			SetBody(&testType{
				Text: "bar",
			}).
			SetResult(replaced).
			Put(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}
		if put.IsError() {
			t.Fatal(put)
		}

		if replaced.Text != "bar" {
			t.Fatalf("%+v", replaced)
		}

		if replaced.Id != created.Id {
			t.Fatalf("%+v %+v", replaced, created)
		}

		read := &testType{}

		_, err = c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}

		if read.Text != "bar" {
			t.Fatalf("%+v", read)
		}

		if read.Num != 0 {
			t.Fatalf("%+v", read)
		}

		if read.Id != created.Id {
			t.Fatalf("%+v %+v", read, created)
		}
	})
}

func TestPUTIfMatch(t *testing.T) {
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
		if etag != fmt.Sprintf(`"%s"`, created.Sha256) {
			t.Fatalf("%s vs %+v", etag, created)
		}

		replaced := &testType{}

		resp, err = c.R().
			SetBody(&testType{
				Text: "bar",
			}).
			SetResult(replaced).
			Put(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}

		etag = resp.Header().Get("ETag")
		if etag != fmt.Sprintf(`"%s"`, replaced.Sha256) {
			t.Fatalf("%s vs %+v", etag, replaced)
		}

		if replaced.Sha256 == created.Sha256 {
			t.Fatalf("sha256 unchanged")
		}

		resp, err = c.R().
			SetHeader("If-Match", `"foobar"`).
			SetBody(&testType{
				Text: "zig",
			}).
			SetResult(replaced).
			Put(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
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
