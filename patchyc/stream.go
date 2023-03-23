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
	ch   <-chan []*T
	body io.ReadCloser

	lastEventReceived time.Time
	lastID            string
	lastError         error

	mu sync.RWMutex
}

func (ls *ListStream[T]) Close() {
	ls.body.Close()
}

func (ls *ListStream[T]) Chan() <-chan []*T {
	return ls.ch
}

func (ls *ListStream[T]) Read() ([]*T, error) {
	// TODO: Send id, time, and error through the channel so we don't get out of sync
	list := <-ls.Chan()

	ls.mu.RLock()
	defer ls.mu.RUnlock()

	return list, ls.lastError
}

func (ls *ListStream[T]) LastEventReceived() time.Time {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	return ls.lastEventReceived
}

func (ls *ListStream[T]) receivedEvent(id string, err error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	ls.lastEventReceived = time.Now()

	if id != "" {
		ls.lastID = id
	}

	if err != nil {
		ls.lastError = err
	}
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
