package patchyc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

type ObjectStream[T any] struct {
	ch   <-chan *T
	body io.ReadCloser
}

func (os *ObjectStream[T]) Close() {
	os.body.Close()
}

func (os *ObjectStream[T]) Chan() <-chan *T {
	return os.ch
}

func (os *ObjectStream[T]) Read() *T {
	return <-os.Chan()
}

func readEvent(scan *bufio.Scanner, out any) (string, error) {
	eventType := ""
	data := [][]byte{}

	for scan.Scan() {
		line := scan.Text()

		switch {
		case strings.HasPrefix(line, ":"):
			continue

		case strings.HasPrefix(line, "event: "):
			eventType = strings.TrimPrefix(line, "event: ")

		case strings.HasPrefix(line, "data: "):
			data = append(data, bytes.TrimPrefix(scan.Bytes(), []byte("data: ")))

		case line == "":
			var err error

			if out != nil {
				err = json.Unmarshal(bytes.Join(data, []byte("\n")), out)
			}

			return eventType, err
		}
	}

	return "", io.EOF
}
