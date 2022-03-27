package potency

import "net/http"

import "github.com/firestuff/patchy/store"

type Potency struct {
	store *store.Store
}

func NewPotency(store *store.Store) *Potency {
	return &Potency{
		store: store,
	}
}

func (p *Potency) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
