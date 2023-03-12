package patchy

import (
	"github.com/firestuff/patchy/api"
)

type (
	API      = api.API
	Metadata = api.Metadata
)

var (
	ErrUnknownAcceptType = api.ErrUnknownAcceptType

	NewFileStoreAPI = api.NewFileStoreAPI
	NewSQLiteAPI    = api.NewSQLiteAPI
	NewAPI          = api.NewAPI
)

func Register[T any](a *API) {
	api.Register[T](a)
}

func RegisterName[T any](a *API, typeName string) {
	api.RegisterName[T](a, typeName)
}
