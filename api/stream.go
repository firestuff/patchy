package api

import (
	"sync"
)

type GetStream[T any] struct {
	ch  chan *GetStreamEvent[T]
	gsi *getStreamInt

	lastID string

	err error

	mu sync.RWMutex
}

type GetStreamEvent[T any] struct {
	ID  string
	Obj *T
}

func (gs *GetStream[T]) Close() {
	gs.gsi.Close()
}

func (gs *GetStream[T]) Chan() <-chan *GetStreamEvent[T] {
	return gs.ch
}

func (gs *GetStream[T]) Read() *GetStreamEvent[T] {
	return <-gs.Chan()
}

func (gs *GetStream[T]) LastID() string {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	return gs.lastID
}

func (gs *GetStream[T]) Error() error {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	return gs.err
}

func (gs *GetStream[T]) writeEvent(id string, obj *T) {
	gs.mu.Lock()
	gs.lastID = id
	gs.mu.Unlock()

	gs.ch <- &GetStreamEvent[T]{
		ID:  id,
		Obj: obj,
	}
}

func (gs *GetStream[T]) writeError(err error) {
	gs.mu.Lock()
	gs.err = err
	gs.mu.Unlock()

	close(gs.ch)
}

type ListStream[T any] struct {
	ch  chan *ListStreamEvent[T]
	lsi *listStreamInt

	lastID string

	err error

	mu sync.RWMutex
}

type ListStreamEvent[T any] struct {
	ID   string
	List []*T
}

func (ls *ListStream[T]) Close() {
	ls.lsi.Close()
}

func (ls *ListStream[T]) Chan() <-chan *ListStreamEvent[T] {
	return ls.ch
}

func (ls *ListStream[T]) Read() *ListStreamEvent[T] {
	return <-ls.Chan()
}

func (ls *ListStream[T]) LastID() string {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	return ls.lastID
}

func (ls *ListStream[T]) Error() error {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	return ls.err
}

func (ls *ListStream[T]) writeEvent(id string, list []*T) {
	ls.mu.Lock()
	ls.lastID = id
	ls.mu.Unlock()

	ls.ch <- &ListStreamEvent[T]{
		ID:   id,
		List: list,
	}
}

func (ls *ListStream[T]) writeError(err error) {
	ls.mu.Lock()
	ls.err = err
	ls.mu.Unlock()

	close(ls.ch)
}
