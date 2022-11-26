package storebus

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/firestuff/patchy/bus"
	"github.com/firestuff/patchy/metadata"
	"github.com/firestuff/patchy/store"
	"github.com/firestuff/patchy/view"
)

type StoreBus struct {
	store store.Storer
	bus   *bus.Bus
	mu    sync.RWMutex
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
		return err
	}

	if err := sb.store.Write(t, obj); err != nil {
		return err
	}

	sb.bus.Announce(t, obj)

	return nil
}

func (sb *StoreBus) Delete(t string, id string) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if err := sb.store.Delete(t, id); err != nil {
		return err
	}

	sb.bus.Delete(t, id)

	return nil
}

func (sb *StoreBus) Read(ctx context.Context, t string, id string, factory func() any) (*view.EphemeralView[any], error) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	obj, err := sb.store.Read(t, id, factory)
	if err != nil {
		return nil, err
	}

	ev, err := view.NewEphemeralView[any](ctx, obj)
	if err != nil {
		return nil, err
	}

	sb.bus.SubscribeKey(t, id, ev)

	return ev, nil
}

func (sb *StoreBus) List(ctx context.Context, t string, factory func() any) (*view.EphemeralView[[]any], error) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	l, err := sb.store.List(t, factory)
	if err != nil {
		return nil, err
	}

	ret, err := view.NewEphemeralView[[]any](ctx, l)
	if err != nil {
		return nil, err
	}

	ev := view.NewEphemeralViewEmpty[any](ctx)
	sb.bus.SubscribeType(t, ev)

	go func() {
		for range ev.Chan() {
			// List() results are always at least (but not exactly) as new as the write that triggered it
			l, err := sb.store.List(t, factory)
			if err != nil {
				break
			}

			err = ret.Update(l)
			if err != nil {
				// Update() closes the channel on failure
				return
			}
		}

		ret.Close()
	}()

	return ret, nil
}

func updateHash(obj any) error {
	m := *metadata.GetMetadata(obj)
	metadata.ClearMetadata(obj)

	defer metadata.SetMetadata(obj, &m)

	hash := sha256.New()
	enc := json.NewEncoder(hash)

	if err := enc.Encode(obj); err != nil {
		return err
	}

	m.ETag = fmt.Sprintf("etag:%x", hash.Sum(nil))

	return nil
}
