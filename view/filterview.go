package view

type FilterView[T any] struct {
	ch chan T
}

func NewFilterView[T any](input ReadView[T], filter func(T) T) *FilterView[T] {
	v := &FilterView[T]{
		ch: make(chan T, 100),
	}

	go func() {
		defer close(v.ch)

		for update := range input.Chan() {
			select {
			case v.ch <- filter(update):

			default:
				break
			}
		}
	}()

	return v
}

func (v *FilterView[T]) Chan() chan T {
	return v.ch
}
