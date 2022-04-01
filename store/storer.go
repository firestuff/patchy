package store

type Storer interface {
	Write(string, any) error
	Delete(string, any) error
	Read(string, any) error
	List(string, func() any) ([]any, error)
}
