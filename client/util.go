package client

func P[T any](v T) *T {
	return &v
}
