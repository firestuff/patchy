package patchy

import (
	"context"

	"github.com/firestuff/patchy/client"
)

type (
	Client = client.Client
)

var (
	NewClient = client.NewClient
	P         = client.P
)

func CreateName[T any](ctx context.Context, c *Client, name string, obj *T) (*T, error) {
	return client.CreateName[T](ctx, c, name, obj)
}

func Create[T any](ctx context.Context, c *Client, obj *T) (*T, error) {
	return client.Create[T](ctx, c, obj)
}

func FindName[T any](ctx context.Context, c *Client, name, shortID string) (*T, error) {
	return client.FindName[T](ctx, c, name, shortID)
}

func Find[T any](ctx context.Context, c *Client, shortID string) (*T, error) {
	return client.Find[T](ctx, c, shortID)
}

func GetName[T any](ctx context.Context, c *Client, name, id string) (*T, error) {
	return client.GetName[T](ctx, c, name, id)
}

func Get[T any](ctx context.Context, c *Client, id string) (*T, error) {
	return client.Get[T](ctx, c, id)
}

func ListName[T any](ctx context.Context, c *Client, name string, opts *ListOpts) ([]*T, error) {
	return client.ListName[T](ctx, c, name, opts)
}

func List[T any](ctx context.Context, c *Client, opts *ListOpts) ([]*T, error) {
	return client.List[T](ctx, c, opts)
}

func ReplaceName[T any](ctx context.Context, c *Client, name, id string, obj *T) (*T, error) {
	return client.ReplaceName[T](ctx, c, name, id, obj)
}

func Replace[T any](ctx context.Context, c *Client, id string, obj *T) (*T, error) {
	return client.Replace[T](ctx, c, id, obj)
}

func UpdateName[T any](ctx context.Context, c *Client, name, id string, obj *T) (*T, error) {
	return client.UpdateName[T](ctx, c, name, id, obj)
}

func Update[T any](ctx context.Context, c *Client, id string, obj *T) (*T, error) {
	return client.Update[T](ctx, c, id, obj)
}
