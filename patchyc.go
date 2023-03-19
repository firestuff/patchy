package patchy

import "github.com/firestuff/patchy/patchyc"

func P[T any](v T) *T {
	return patchyc.P(v)
}
