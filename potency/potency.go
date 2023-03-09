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

	Method        string      `json:"method"`
	URL           string      `json:"url"`
	RequestHeader http.Header `json:"requestHeader"`
	Sha256        string      `json:"sha256"`

	StatusCode     int         `json:"statusCode"`
	ResponseHeader http.Header `json:"responseHeader"`
	ResponseBody   []byte      `json:"responseBody"`
}

var (
	ErrConflict       = errors.New("conflict")
	ErrMismatch       = errors.New("idempotency mismatch")
	ErrBodyMismatch   = fmt.Errorf("request body mismatch: %w", ErrMismatch)
	ErrMethodMismatch = fmt.Errorf("HTTP method mismatch: %w", ErrMismatch)
	ErrURLMismatch    = fmt.Errorf("URL mismatch: %w", ErrMismatch)
	ErrHeaderMismatch = fmt.Errorf("Header mismatch: %w", ErrMismatch)
	ErrInvalidKey     = errors.New("invalid Idempotency-Key")

	criticalHeaders = []string{
		"Accept",
		"Authorization",
	}
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
		err := jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", val, ErrInvalidKey)
		jsrest.WriteError(w, err)
		return
	}

	key := val[1 : len(val)-1]

	rd, err := p.store.Read("idempotency-key", key, func() any { return &savedResult{} })
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrInternalServerError, "read idempotency key failed: %s (%w)", key, err)
		jsrest.WriteError(w, err)
		return
	}

	if rd != nil {
		saved := rd.(*savedResult)

		if r.Method != saved.Method {
			err = jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", r.Method, ErrMethodMismatch)
			jsrest.WriteError(w, err)
			return
		}

		if r.URL.String() != saved.URL {
			err = jsrest.Errorf(jsrest.ErrBadRequest, "%s (%w)", r.URL.String(), ErrURLMismatch)
			jsrest.WriteError(w, err)
			return
		}

		for _, h := range criticalHeaders {
			if saved.RequestHeader.Get(h) != r.Header.Get(h) {
				err = jsrest.Errorf(jsrest.ErrBadRequest, "%s: %s (%w)", h, r.Header.Get(h), ErrHeaderMismatch)
				jsrest.WriteError(w, err)
				return
			}
		}

		h := sha256.New()

		_, err = io.Copy(h, r.Body)
		if err != nil {
			err = jsrest.Errorf(jsrest.ErrBadRequest, "hash request body failed (%w)", err)
			jsrest.WriteError(w, err)
			return
		}

		hexed := hex.EncodeToString(h.Sum(nil))
		if hexed != saved.Sha256 {
			err = jsrest.Errorf(jsrest.ErrUnprocessableEntity, "%s vs %s (%w)", hexed, saved.Sha256, ErrBodyMismatch)
			jsrest.WriteError(w, err)
			return
		}

		for key, vals := range saved.ResponseHeader {
			w.Header().Set(key, vals[0])
		}

		w.WriteHeader(saved.StatusCode)
		_, _ = w.Write(saved.ResponseBody)

		return
	}

	// Store miss, proceed to normal execution with interception
	err = p.lockKey(key)
	if err != nil {
		err = jsrest.Errorf(jsrest.ErrConflict, "%s", key)
		jsrest.WriteError(w, err)
		return
	}

	defer p.unlockKey(key)

	requestHeader := http.Header{}
	for _, h := range criticalHeaders {
		requestHeader.Set(h, r.Header.Get(h))
	}

	bi := newBodyIntercept(r.Body)
	r.Body = bi

	rwi := newResponseWriterIntercept(w)
	w = rwi

	p.handler.ServeHTTP(w, r)

	save := &savedResult{
		Metadata: metadata.Metadata{
			ID: key,
		},

		Method:        r.Method,
		URL:           r.URL.String(),
		RequestHeader: requestHeader,
		Sha256:        hex.EncodeToString(bi.sha256.Sum(nil)),

		StatusCode:     rwi.statusCode,
		ResponseHeader: rwi.Header(),
		ResponseBody:   rwi.buf.Bytes(),
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
