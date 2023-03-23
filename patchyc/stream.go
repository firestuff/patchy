package patchyc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

type GetStream[T any] struct {
	ch   <-chan *T
	body io.ReadCloser
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
}

func (ls *ListStream[T]) Close() {
	ls.body.Close()
}

func (ls *ListStream[T]) Chan() <-chan []*T {
	return ls.ch
}

func (ls *ListStream[T]) Read() []*T {
	return <-ls.Chan()
}

func readEvent(scan *bufio.Scanner, out any) (string, string, error) { //nolint:unparam
	// TODO: Remove unparam above
	eventType := ""
	id := ""
	data := [][]byte{}

	for scan.Scan() {
		line := scan.Text()

		switch {
		case strings.HasPrefix(line, ":"):
			continue

		case strings.HasPrefix(line, "event: "):
			eventType = strings.TrimPrefix(line, "event: ")

		case strings.HasPrefix(line, "id: "):
			id = strings.TrimPrefix(line, "id: ")

		case strings.HasPrefix(line, "data: "):
			data = append(data, bytes.TrimPrefix(scan.Bytes(), []byte("data: ")))

		case line == "":
			var err error

			if out != nil {
				err = json.Unmarshal(bytes.Join(data, []byte("\n")), out)
			}

			return eventType, id, err
		}
	}

	return "", "", io.EOF
}
