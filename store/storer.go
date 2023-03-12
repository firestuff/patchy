package store

import "context"

type Storer interface {
	Close()
	Write(context.Context, string, any) error
	Delete(context.Context, string, string) error
	Read(context.Context, string, string, func() any) (any, error)
	List(context.Context, string, func() any) ([]any, error)
}
