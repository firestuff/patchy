package storebus

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/firestuff/patchy/bus"
	"github.com/firestuff/patchy/metadata"
	"github.com/firestuff/patchy/store"
)

type StoreBus struct {
	store store.Storer
	bus   *bus.Bus
}

func NewStoreBus(st store.Storer) *StoreBus {
	return &StoreBus{
		store: st,
		bus:   bus.NewBus(),
	}
}

func (sb *StoreBus) Write(t string, obj any) error {
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
	if err := sb.store.Delete(t, id); err != nil {
		return err
	}

	sb.bus.Delete(t, id)

	return nil
}

func (sb *StoreBus) Read(t string, obj any) error {
	return sb.store.Read(t, obj)
}

func (sb *StoreBus) List(t string, factory func() any) ([]any, error) {
	return sb.store.List(t, factory)
}

func (sb *StoreBus) SubscribeKey(t string, id string) chan any {
	return sb.bus.SubscribeKey(t, id)
}

func (sb *StoreBus) SubscribeType(t string) (chan any, chan string) {
	return sb.bus.SubscribeType(t)
}

func (sb *StoreBus) GetBus() *bus.Bus {
	return sb.bus
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
