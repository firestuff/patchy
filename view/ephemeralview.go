package view

import (
	"context"
	"fmt"
)

type EphemeralView[T any] struct {
	ctx    context.Context
	ch     chan T
	closed bool
}

func NewEphemeralView[T any](ctx context.Context, data T) (*EphemeralView[T], error) {
	v := NewEphemeralViewEmpty[T](ctx)

	if err := v.Update(data); err != nil {
		return nil, err
	}

	return v, nil
}

func NewEphemeralViewEmpty[T any](ctx context.Context) *EphemeralView[T] {
	return &EphemeralView[T]{
		ctx:    ctx,
		ch:     make(chan T, 100),
		closed: false,
	}
}

var errChannelOverrun = fmt.Errorf("channel overrun")

func (v *EphemeralView[T]) Chan() <-chan T {
	return v.ch
}

func (v *EphemeralView[T]) Update(data T) error {
	select {
	case <-v.ctx.Done():
		close(v.ch)
		return v.ctx.Err()

	case v.ch <- data:
		return nil

	default:
		close(v.ch)
		return errChannelOverrun
	}
}

func (v *EphemeralView[T]) MustUpdate(data T) {
	if err := v.Update(data); err != nil {
		panic(err)
	}
}

func (v *EphemeralView[T]) Close() {
	close(v.ch)
}
