package storebus

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/firestuff/patchy/bus"
	"github.com/firestuff/patchy/jsrest"
	"github.com/firestuff/patchy/metadata"
	"github.com/firestuff/patchy/store"
)

type StoreBus struct {
	store store.Storer
	bus   *bus.Bus

	// This lock ensures that no writes interleave with read/subscribe pairs
	mu sync.RWMutex
}

func NewStoreBus(st store.Storer) *StoreBus {
	return &StoreBus{
		store: st,
		bus:   bus.NewBus(),
	}
}

func (sb *StoreBus) Write(t string, obj any) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if err := updateHash(obj); err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "hash update failed (%w)", err)
	}

	if err := sb.store.Write(t, obj); err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "write failed (%w)", err)
	}

	sb.bus.Announce(t, obj)

	return nil
}

func (sb *StoreBus) Delete(t, id string) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if err := sb.store.Delete(t, id); err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "delete failed (%w)", err)
	}

	sb.bus.Delete(t, id)

	return nil
}

func (sb *StoreBus) Read(t, id string, factory func() any) (any, error) {
	return sb.store.Read(t, id, factory)
}

func (sb *StoreBus) ReadStream(t, id string, factory func() any) (<-chan any, error) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	initial, err := sb.store.Read(t, id, factory)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "read failed (%w)", err)
	}

	c := sb.bus.SubscribeKey(t, id, initial)

	return c, nil
}

func (sb *StoreBus) CloseReadStream(t, id string, c <-chan any) {
	sb.bus.UnsubscribeKey(t, id, c)
}

func (sb *StoreBus) List(t string, factory func() any) ([]any, error) {
	return sb.store.List(t, factory)
}

func (sb *StoreBus) ListStream(t string, factory func() any) (<-chan []any, error) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	initial, err := sb.store.List(t, factory)
	if err != nil {
		return nil, jsrest.Errorf(jsrest.ErrInternalServerError, "list failed (%w)", err)
	}

	c := sb.bus.SubscribeType(t, initial)

	ret := make(chan []any, 100)

	go func() {
		defer close(ret)
		defer sb.bus.UnsubscribeType(t, c)

		for range c {
			// List() results are always at least (but not exactly) as new as the write that triggered it
			l, err := sb.store.List(t, factory)
			if err != nil {
				break
			}

			select {
			case ret <- l:
			default:
				break
			}
		}
	}()

	return ret, nil
}

func (sb *StoreBus) CloseListStream(t string, c <-chan any) {
	sb.bus.UnsubscribeType(t, c)
}

func updateHash(obj any) error {
	m := *metadata.GetMetadata(obj)
	metadata.ClearMetadata(obj)

	defer metadata.SetMetadata(obj, &m)

	hash := sha256.New()
	enc := json.NewEncoder(hash)

	if err := enc.Encode(obj); err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "JSON encode failed (%w)", err)
	}

	m.ETag = fmt.Sprintf("etag:%x", hash.Sum(nil))

	return nil
}
