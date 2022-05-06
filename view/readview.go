package view

type ReadView[T any] interface {
	Chan() chan T
}
