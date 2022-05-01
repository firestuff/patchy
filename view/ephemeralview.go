package view

import "fmt"

type EphemeralView[T any] struct {
	ch     chan T
	closed bool
}

func NewEphemeralView[T any](data T) *EphemeralView[T] {
	v := &EphemeralView[T]{
		ch:     make(chan T, 100),
		closed: false,
	}

	v.MustUpdate(data)

	return v
}

var errChannelOverrun = fmt.Errorf("channel overrun")

func (v *EphemeralView[T]) Chan() chan T {
	return v.ch
}

func (v *EphemeralView[T]) Close() {
	if v.closed {
		return
	}

	close(v.ch)
	v.closed = true
}

func (v *EphemeralView[T]) Update(data T) error {
	select {
	case v.ch <- data:
		return nil

	default:
		v.Close()
		return errChannelOverrun
	}
}

func (v *EphemeralView[T]) MustUpdate(data T) {
	if err := v.Update(data); err != nil {
		panic(err)
	}
}
