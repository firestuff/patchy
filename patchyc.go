package patchy

import "github.com/gopatchy/patchyc"

func P[T any](v T) *T {
	return patchyc.P(v)
}
