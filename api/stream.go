package api

type ObjectStream[T any] struct {
	ch  <-chan *T
	ios *intObjectStream
}

func (os *ObjectStream[T]) Close() {
	os.ios.Close()
}

func (os *ObjectStream[T]) Chan() <-chan *T {
	return os.ch
}

func (os *ObjectStream[T]) Read() *T {
	return <-os.Chan()
}

type ObjectListStream[T any] struct {
	ch   <-chan []*T
	iols *intObjectListStream
}

func (ols *ObjectListStream[T]) Close() {
	ols.iols.Close()
}

func (ols *ObjectListStream[T]) Chan() <-chan []*T {
	return ols.ch
}

func (ols *ObjectListStream[T]) Read() []*T {
	return <-ols.Chan()
}
