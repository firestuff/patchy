package patchy

import (
	"context"

	"github.com/firestuff/patchy/api"
)

type (
	API         = api.API
	DebugInfo   = api.DebugInfo
	Filter      = api.Filter
	GetOpts     = api.GetOpts
	ListHook    = api.ListHook
	ListOpts    = api.ListOpts
	Metadata    = api.Metadata
	OpenAPI     = api.OpenAPI
	OpenAPIInfo = api.OpenAPIInfo
	UpdateOpts  = api.UpdateOpts
)

var (
	ErrUnknownAcceptType = api.ErrUnknownAcceptType

	NewFileStoreAPI = api.NewFileStoreAPI
	NewSQLiteAPI    = api.NewSQLiteAPI
	NewAPI          = api.NewAPI
)

const (
	ContextInternal = api.ContextInternal

	ContextAuthBasic  = api.ContextAuthBasic
	ContextAuthBearer = api.ContextAuthBearer
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

func DeleteName[T any](ctx context.Context, a *API, name, id string, opts *UpdateOpts) error {
	return api.DeleteName[T](ctx, a, name, id, opts)
}

func Delete[T any](ctx context.Context, a *API, id string, opts *UpdateOpts) error {
	return api.Delete[T](ctx, a, id, opts)
}

func FindName[T any](ctx context.Context, a *API, name, shortID string) (*T, error) {
	return api.FindName[T](ctx, a, name, shortID)
}

func Find[T any](ctx context.Context, a *API, shortID string) (*T, error) {
	return api.Find[T](ctx, a, shortID)
}

func GetName[T any](ctx context.Context, a *API, name, id string, opts *GetOpts) (*T, error) {
	return api.GetName[T](ctx, a, name, id, opts)
}

func Get[T any](ctx context.Context, a *API, id string, opts *GetOpts) (*T, error) {
	return api.Get[T](ctx, a, id, opts)
}

func ListName[T any](ctx context.Context, a *API, name string, opts *ListOpts) ([]*T, error) {
	return api.ListName[T](ctx, a, name, opts)
}

func List[T any](ctx context.Context, a *API, opts *ListOpts) ([]*T, error) {
	return api.List[T](ctx, a, opts)
}

func ReplaceName[T any](ctx context.Context, a *API, name, id string, obj *T, opts *UpdateOpts) (*T, error) {
	return api.ReplaceName[T](ctx, a, name, id, obj, opts)
}

func Replace[T any](ctx context.Context, a *API, id string, obj *T, opts *UpdateOpts) (*T, error) {
	return api.Replace[T](ctx, a, id, obj, opts)
}

func UpdateName[T any](ctx context.Context, a *API, name, id string, obj *T, opts *UpdateOpts) (*T, error) {
	return api.UpdateName[T](ctx, a, name, id, obj, opts)
}

func Update[T any](ctx context.Context, a *API, id string, obj *T, opts *UpdateOpts) (*T, error) {
	return api.Update[T](ctx, a, id, obj, opts)
}

func SetAuthBasicName[T any](a *API, name, pathUser, pathPass string) {
	api.SetAuthBasicName[T](a, name, pathUser, pathPass)
}

func SetAuthBasic[T any](a *API, pathUser, pathPass string) {
	api.SetAuthBasic[T](a, pathUser, pathPass)
}

func SetAuthBearerName[T any](a *API, name, pathToken string) {
	api.SetAuthBearerName[T](a, name, pathToken)
}

func SetAuthBearer[T any](a *API, pathToken string) {
	api.SetAuthBearer[T](a, pathToken)
}

func SetListHookName[T any](a *API, name string, hook ListHook) {
	api.SetListHookName[T](a, name, hook)
}

func SetListHook[T any](a *API, hook ListHook) {
	api.SetListHook[T](a, hook)
}

func IsCreate[T any](obj *T, prev *T) bool {
	return api.IsCreate[T](obj, prev)
}

func IsUpdate[T any](obj *T, prev *T) bool {
	return api.IsUpdate[T](obj, prev)
}

func IsDelete[T any](obj *T, prev *T) bool {
	return api.IsDelete[T](obj, prev)
}

func FieldChanged[T any](obj *T, prev *T, p string) bool {
	return api.FieldChanged[T](obj, prev, p)
}
