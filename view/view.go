package view

type View[T any] interface {
	Chan() chan T
}
