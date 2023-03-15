package patchy

import (
	"context"

	"github.com/firestuff/patchy/api"
)

type (
	API      = api.API
	Metadata = api.Metadata
	ListOpts = api.ListOpts
	Filter   = api.Filter
)

var (
	ErrUnknownAcceptType = api.ErrUnknownAcceptType

	NewFileStoreAPI = api.NewFileStoreAPI
	NewSQLiteAPI    = api.NewSQLiteAPI
	NewAPI          = api.NewAPI

	DeleteName = api.DeleteName
)

func Register[T any](a *API) {
	api.Register[T](a)
}

func RegisterName[T any](a *API, typeName string) {
	api.RegisterName[T](a, typeName)
}

func CreateName[T any](ctx context.Context, a *API, name string, obj *T) (*T, error) {
	return api.CreateName[T](ctx, a, name, obj)
}

func Create[T any](ctx context.Context, a *API, obj *T) (*T, error) {
	return api.Create[T](ctx, a, obj)
}

func Delete[T any](ctx context.Context, a *API, id string) error {
	return api.Delete[T](ctx, a, id)
}

func FindName[T any](ctx context.Context, a *API, name, shortID string) (*T, error) {
	return api.FindName[T](ctx, a, name, shortID)
}

func Find[T any](ctx context.Context, a *API, shortID string) (*T, error) {
	return api.Find[T](ctx, a, shortID)
}

func GetName[T any](ctx context.Context, a *API, name, id string) (*T, error) {
	return api.GetName[T](ctx, a, name, id)
}

func Get[T any](ctx context.Context, a *API, id string) (*T, error) {
	return api.Get[T](ctx, a, id)
}

func ListName[T any](ctx context.Context, a *API, name string, opts *ListOpts) ([]*T, error) {
	return api.ListName[T](ctx, a, name, opts)
}

func List[T any](ctx context.Context, a *API, opts *ListOpts) ([]*T, error) {
	return api.List[T](ctx, a, opts)
}

func ReplaceName[T any](ctx context.Context, a *API, name, id string, obj *T) (*T, error) {
	return api.ReplaceName[T](ctx, a, name, id, obj)
}

func Replace[T any](ctx context.Context, a *API, id string, obj *T) (*T, error) {
	return api.Replace[T](ctx, a, id, obj)
}

func UpdateName[T any](ctx context.Context, a *API, name, id string, obj *T) (*T, error) {
	return api.UpdateName[T](ctx, a, name, id, obj)
}

func Update[T any](ctx context.Context, a *API, id string, obj *T) (*T, error) {
	return api.Update[T](ctx, a, id, obj)
}
