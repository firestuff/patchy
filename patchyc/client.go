package patchyc

import (
	"context"
	"crypto/tls"

	"github.com/firestuff/patchy/api"
	"github.com/firestuff/patchy/jsrest"
	"github.com/go-resty/resty/v2"
)

// TODO: Add retries
// TODO: Add Idempotency-Key support

type (
	DebugInfo = api.DebugInfo
	OpenAPI   = api.OpenAPI
)

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

func (c *Client) ResetAuth() *Client {
	c.rst.Token = ""
	c.rst.UserInfo = nil

	return c
}

func (c *Client) SetBasicAuth(user, pass string) *Client {
	c.ResetAuth()
	c.rst.SetBasicAuth(user, pass)

	return c
}

func (c *Client) SetAuthToken(token string) *Client {
	c.ResetAuth()
	c.rst.SetAuthToken(token)

	return c
}

func (c *Client) SetHeader(header, value string) *Client {
	c.rst.SetHeader(header, value)
	return c
}

func (c *Client) DebugInfo(ctx context.Context) (*DebugInfo, error) {
	ret := &DebugInfo{}

	resp, err := c.rst.R().
		SetContext(ctx).
		SetResult(ret).
		Get("_debug")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, jsrest.ReadError(resp.Body())
	}

	return ret, nil
}

func (c *Client) OpenAPI(ctx context.Context) (*OpenAPI, error) {
	ret := &OpenAPI{}

	resp, err := c.rst.R().
		SetContext(ctx).
		SetResult(ret).
		Get("_openapi")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, jsrest.ReadError(resp.Body())
	}

	return ret, nil
}

func (c *Client) GoClient(ctx context.Context) (string, error) {
	resp, err := c.rst.R().
		SetContext(ctx).
		Get("_client.go")
	if err != nil {
		return "", err
	}

	if resp.IsError() {
		return "", jsrest.ReadError(resp.Body())
	}

	return resp.String(), nil
}

func (c *Client) TSClient(ctx context.Context) (string, error) {
	resp, err := c.rst.R().
		SetContext(ctx).
		Get("_client.ts")
	if err != nil {
		return "", err
	}

	if resp.IsError() {
		return "", jsrest.ReadError(resp.Body())
	}

	return resp.String(), nil
}
