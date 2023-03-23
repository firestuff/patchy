package patchyc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"time"
)

type GetStream[T any] struct {
	ch   <-chan *T
	body io.ReadCloser

	// TODO: Make GetStream match ListStream
}

func (gs *GetStream[T]) Close() {
	gs.body.Close()
}

func (gs *GetStream[T]) Chan() <-chan *T {
	return gs.ch
}

func (gs *GetStream[T]) Read() *T {
	return <-gs.Chan()
}

type ListStream[T any] struct {
	ch   chan *ListStreamEvent[T]
	body io.ReadCloser

	lastEventReceived time.Time
	lastID            string

	err error

	mu sync.RWMutex
}

type ListStreamEvent[T any] struct {
	ID   string
	List []*T
}

func (ls *ListStream[T]) Close() {
	ls.body.Close()
}

func (ls *ListStream[T]) Chan() <-chan *ListStreamEvent[T] {
	return ls.ch
}

func (ls *ListStream[T]) Read() *ListStreamEvent[T] {
	return <-ls.Chan()
}

func (ls *ListStream[T]) LastEventReceived() time.Time {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	return ls.lastEventReceived
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

func (ls *ListStream[T]) writeHeartbeat() {
	ls.mu.Lock()
	ls.lastEventReceived = time.Now()
	ls.mu.Unlock()
}

func (ls *ListStream[T]) writeEvent(id string, list []*T) {
	ls.mu.Lock()
	ls.lastEventReceived = time.Now()
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

type streamEvent struct {
	eventType string
	id        string
	data      []byte
}

func readEvent(scan *bufio.Scanner) (*streamEvent, error) {
	event := &streamEvent{}
	data := [][]byte{}

	for scan.Scan() {
		line := scan.Text()

		switch {
		case strings.HasPrefix(line, ":"):
			continue

		case strings.HasPrefix(line, "event: "):
			event.eventType = strings.TrimPrefix(line, "event: ")

		case strings.HasPrefix(line, "id: "):
			event.id = strings.TrimPrefix(line, "id: ")

		case strings.HasPrefix(line, "data: "):
			data = append(data, bytes.TrimPrefix(scan.Bytes(), []byte("data: ")))

		case line == "":
			event.data = bytes.Join(data, []byte("\n"))
			return event, nil
		}
	}

	return nil, io.EOF
}

func (event *streamEvent) decode(out any) error {
	return json.Unmarshal(event.data, out)
}
