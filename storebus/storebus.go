package storebus

import "crypto/sha256"
import "encoding/hex"
import "encoding/json"

import "github.com/firestuff/patchy/bus"
import "github.com/firestuff/patchy/metadata"
import "github.com/firestuff/patchy/store"

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

func (sb *StoreBus) Write(t string, obj interface{}) error {
	err := updateHash(obj)
	if err != nil {
		return err
	}

	err = sb.store.Write(t, obj)
	if err != nil {
		return err
	}

	sb.bus.Announce(t, obj)

	return nil
}

func (sb *StoreBus) Delete(t string, obj interface{}) error {
	err := sb.store.Delete(t, obj)
	if err != nil {
		return err
	}

	sb.bus.Delete(t, obj)

	return nil
}

func (sb *StoreBus) Read(t string, obj interface{}) error {
	return sb.store.Read(t, obj)
}

func (sb *StoreBus) List(t string, factory func() any) ([]any, error) {
	return sb.store.List(t, factory)
}

func (sb *StoreBus) Subscribe(t string, obj interface{}) chan interface{} {
	return sb.bus.SubscribeKey(t, obj)
}

func (sb *StoreBus) GetStore() store.Storer {
	return sb.store
}

func (sb *StoreBus) GetBus() *bus.Bus {
	return sb.bus
}

func updateHash(obj interface{}) error {
	m := metadata.GetMetadata(obj)
	m.ETag = ""

	hash := sha256.New()
	enc := json.NewEncoder(hash)

	err := enc.Encode(obj)
	if err != nil {
		return err
	}

	m.ETag = hex.EncodeToString(hash.Sum(nil))

	return nil
}
