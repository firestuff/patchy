package potency

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/firestuff/patchy/metadata"
	"github.com/firestuff/patchy/store"
)

type Potency struct {
	store store.Storer

	inProgress map[string]bool
	mu         sync.Mutex
}

type savedResult struct {
	metadata.Metadata

	Method string `json:"method"`
	URL    string `json:"url"`
	Sha256 string `json:"sha256"`

	StatusCode int         `json:"statusCode"`
	Header     http.Header `json:"header"`
	Result     []byte      `json:"result"`
}

func NewPotency(store store.Storer) *Potency {
	return &Potency{
		store:      store,
		inProgress: map[string]bool{},
	}
}

func (p *Potency) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Header.Get("Idempotency-Key")
		if val == "" {
			next.ServeHTTP(w, r)
			return
		}

		if len(val) < 2 || !strings.HasPrefix(val, `"`) || !strings.HasSuffix(val, `"`) {
			http.Error(w, "Invalid Idempotency-Key", http.StatusBadRequest)
			return
		}

		key := val[1 : len(val)-1]

		saved := &savedResult{
			Metadata: metadata.Metadata{
				Id: key,
			},
		}

		err := p.store.Read("idempotency-key", saved)
		if err == nil {
			if r.Method != saved.Method {
				http.Error(w, "Idempotency-Key method mismatch", http.StatusBadRequest)
				return
			}

			if r.URL.String() != saved.URL {
				http.Error(w, "Idempotency-Key URL mismatch", http.StatusBadRequest)
				return
			}

			h := sha256.New()
			_, err = io.Copy(h, r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if hex.EncodeToString(h.Sum(nil)) != saved.Sha256 {
				http.Error(w, "Idempotency-Key request body mismatch", http.StatusUnprocessableEntity)
				return
			}

			// TODO: Ability to verify specified headers match (e.g. authentication token) before returning

			w.WriteHeader(saved.StatusCode)
			_, _ = w.Write(saved.Result)
			return
		}

		// Store miss, proceed to normal execution with interception
		err = p.lockKey(key)
		if err != nil {
			http.Error(w, "Conflict", http.StatusConflict)
			return
		}
		defer p.unlockKey(key)

		bi := newBodyIntercept(r.Body)
		r.Body = bi

		rwi := newResponseWriterIntercept(w)
		w = rwi

		next.ServeHTTP(w, r)

		save := &savedResult{
			Metadata: metadata.Metadata{
				Id: key,
			},

			Method: r.Method,
			URL:    r.URL.String(),
			Sha256: hex.EncodeToString(bi.sha256.Sum(nil)),

			StatusCode: rwi.statusCode,
			Header:     rwi.Header(),
			Result:     rwi.buf.Bytes(),
		}

		// Can't really do anything with the error
		_ = p.store.Write("idempotency-key", save)
	})
}

func (p *Potency) lockKey(key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.inProgress[key] {
		return fmt.Errorf("Conflict")
	}

	p.inProgress[key] = true

	return nil
}

func (p *Potency) unlockKey(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.inProgress, key)
}
