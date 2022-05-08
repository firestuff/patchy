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

	ev := view.NewEphemeralView[any](ctx, obj)

	sb.bus.SubscribeKey(t, id, ev)

	return ev, nil
}

func (sb *StoreBus) List(t string, factory func() any) ([]any, error) {
	// TODO: RLock
	// TODO: Combine with SubscribeType
	return sb.store.List(t, factory)
}

func (sb *StoreBus) SubscribeType(t string) (chan any, chan string) {
	return sb.bus.SubscribeType(t)
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
