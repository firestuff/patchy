package api

import "fmt"
import "testing"

import "github.com/go-resty/resty/v2"

func TestPOST(t *testing.T) {
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

		if created.Text != "foo" {
			t.Fatalf("%s", created.Text)
		}

		if created.Id == "" {
			t.Fatalf("empty Id")
		}

		read := &testType{}

		_, err = c.R().
			SetResult(read).
			Get(fmt.Sprintf("%s/testtype/%s", baseURL, created.Id))
		if err != nil {
			t.Fatal(err)
		}

		if read.Text != "foo" {
			t.Fatalf("%s", read.Text)
		}

		if read.Id != read.Id {
			t.Fatalf("%s %s", read.Id, created.Id)
		}
	})
}
