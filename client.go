package patchy

import "github.com/firestuff/patchy/client"

func P[T any](v T) *T {
	return client.P(v)
}
