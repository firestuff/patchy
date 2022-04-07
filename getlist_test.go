package api

import "fmt"
import "testing"

import "github.com/go-resty/resty/v2"

func TestGETList(t *testing.T) {
	t.Parallel()

	withAPI(t, func(t *testing.T, api *API, baseURL string, c *resty.Client) {
		created1 := &testType{}

		resp, err := c.R().
			SetBody(&testType{
				Text: "foo",
			}).
			SetResult(created1).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		if err != nil {
			t.Fatal(err)
		}
		if resp.IsError() {
			t.Fatal(resp)
		}

		created2 := &testType{}

		resp, err = c.R().
			SetBody(&testType{
				Text: "bar",
			}).
			SetResult(created2).
			Post(fmt.Sprintf("%s/testtype", baseURL))
		if err != nil {
			t.Fatal(err)
		}
		if resp.IsError() {
			t.Fatal(resp)
		}

		list := []testType{}

		resp, err = c.R().
			SetResult(&list).
			Get(fmt.Sprintf("%s/testtype", baseURL))
		if err != nil {
			t.Fatal(err)
		}
		if resp.IsError() {
			t.Fatal(resp)
		}

		if len(list) != 2 {
			t.Fatalf("%+v", list)
		}

		if !((list[0].Text == "foo" && list[1].Text == "bar") ||
			(list[0].Text == "bar" && list[1].Text == "foo")) {
			t.Fatalf("%+v", list)
		}
	})
}
