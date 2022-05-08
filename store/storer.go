package store

type Storer interface {
	Write(string, any) error
	Delete(string, string) error
	Read(string, string, func() any) (any, error)
	List(string, func() any) ([]any, error)
}
