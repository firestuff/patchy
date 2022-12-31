package potency

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
	"github.com/firestuff/patchy/store"
)

type Potency struct {
	store   store.Storer
	handler http.Handler

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

var (
	ErrConflict       = errors.New("conflict")
	ErrMismatch       = errors.New("idempotency mismatch")
	ErrBodyMismatch   = fmt.Errorf("request body mismatch: %w", ErrMismatch)
	ErrMethodMismatch = fmt.Errorf("HTTP method mismatch: %w", ErrMismatch)
	ErrURLMismatch    = fmt.Errorf("URL mismatch: %w", ErrMismatch)
	ErrInvalidKey     = errors.New("invalid Idempotency-Key")
)

func NewPotency(store store.Storer, handler http.Handler) *Potency {
	return &Potency{
		store:      store,
		handler:    handler,
		inProgress: map[string]bool{},
	}
}

func (p *Potency) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	val := r.Header.Get("Idempotency-Key")
	if val == "" {
		p.handler.ServeHTTP(w, r)
		return
	}

	if len(val) < 2 || !strings.HasPrefix(val, `"`) || !strings.HasSuffix(val, `"`) {
		e := fmt.Errorf("%s: %w", val, ErrInvalidKey)
		jse := jsrest.FromError(e, jsrest.StatusBadRequest)
		jse.Write(w)

		return
	}

	key := val[1 : len(val)-1]

	rd, err := p.store.Read("idempotency-key", key, func() any { return &savedResult{} })
	if err != nil {
		e := fmt.Errorf("failed to read idempotency key %s: %w", key, err)
		jse := jsrest.FromError(e, jsrest.StatusInternalServerError)
		jse.Write(w)

		return
	}

	if rd != nil {
		saved := rd.(*savedResult)

		if r.Method != saved.Method {
			e := fmt.Errorf("%s vs %s: %w", r.Method, saved.Method, ErrMethodMismatch)
			jse := jsrest.FromError(e, jsrest.StatusBadRequest)
			jse.Write(w)

			return
		}

		if r.URL.String() != saved.URL {
			e := fmt.Errorf("%s vs %s: %w", r.URL.String(), saved.URL, ErrURLMismatch)
			jse := jsrest.FromError(e, jsrest.StatusBadRequest)
			jse.Write(w)

			return
		}

		h := sha256.New()

		_, err = io.Copy(h, r.Body)
		if err != nil {
			e := fmt.Errorf("failed to hash request body: %w", err)
			jse := jsrest.FromError(e, jsrest.StatusBadRequest)
			jse.Write(w)

			return
		}

		hexed := hex.EncodeToString(h.Sum(nil))
		if hexed != saved.Sha256 {
			e := fmt.Errorf("%s vs %s: %w", hexed, saved.Sha256, ErrBodyMismatch)
			jse := jsrest.FromError(e, jsrest.StatusUnprocessableEntity)
			jse.Write(w)

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
		jse := jsrest.FromError(err, jsrest.StatusConflict)
		jse.Write(w)

		return
	}

	defer p.unlockKey(key)

	bi := newBodyIntercept(r.Body)
	r.Body = bi

	rwi := newResponseWriterIntercept(w)
	w = rwi

	p.handler.ServeHTTP(w, r)

	save := &savedResult{
		Metadata: metadata.Metadata{
			ID: key,
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
}

func (p *Potency) lockKey(key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.inProgress[key] {
		return ErrConflict
	}

	p.inProgress[key] = true

	return nil
}

func (p *Potency) unlockKey(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.inProgress, key)
}
