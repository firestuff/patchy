package api

type GetStream[T any] struct {
	ch  <-chan *T
	gsi *getStreamInt
}

func (gs *GetStream[T]) Close() {
	gs.gsi.Close()
}

func (gs *GetStream[T]) Chan() <-chan *T {
	return gs.ch
}

func (gs *GetStream[T]) Read() *T {
	return <-gs.Chan()
}

type ListStream[T any] struct {
	ch  <-chan []*T
	lsi *listStreamInt
}

func (ls *ListStream[T]) Close() {
	ls.lsi.Close()
}

func (ls *ListStream[T]) Chan() <-chan []*T {
	return ls.ch
}

func (ls *ListStream[T]) Read() []*T {
	return <-ls.Chan()
}
