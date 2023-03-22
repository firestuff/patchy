package patchyc

import (
	"crypto/tls"

	"github.com/go-resty/resty/v2"
)

// TODO: Add retries
// TODO: Add Idempotency-Key support

type Client struct {
	rst *resty.Client
}

func NewClient(baseURL string) *Client {
	rst := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Accept", "application/json").
		SetJSONEscapeHTML(false)

	return &Client{
		rst: rst,
	}
}

func (c *Client) SetTLSClientConfig(cfg *tls.Config) *Client {
	c.rst.SetTLSClientConfig(cfg)
	return c
}

func (c *Client) SetDebug(debug bool) *Client {
	c.rst.SetDebug(debug)
	return c
}

func (c *Client) SetBasicAuth(user, pass string) *Client {
	c.rst.SetBasicAuth(user, pass)
	return c
}

func (c *Client) SetAuthToken(token string) *Client {
	c.rst.SetAuthToken(token)
	return c
}

func (c *Client) SetHeader(header, value string) *Client {
	c.rst.SetHeader(header, value)
	return c
}
