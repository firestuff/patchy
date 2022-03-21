package storebus

import "sync"

type StoreBus struct {
	store *Store
	bus   *Bus

	mu sync.Mutex
}

func NewStoreBus(root string) *StoreBus {
	return &StoreBus{
		store: NewStore(root),
		bus:   NewBus(),
	}
}

func (sb *StoreBus) Write(obj Object) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	err := sb.store.Write(obj)
	if err != nil {
		return err
	}

	sb.bus.Announce(obj)

	return nil
}

func (sb *StoreBus) Read(obj Object) error {
	return sb.store.Read(obj)
}

func (sb *StoreBus) Subscribe(obj Object) chan Object {
	// This mutex prevents Announce() in another goroutine after we call
	// Subscribe() but before we write to the channel. That would cause the
	// objects in the channel to be out of order.

	sb.mu.Lock()
	defer sb.mu.Unlock()

	ch := sb.bus.Subscribe(obj)

	err := sb.store.Read(obj)
	if err == nil {
		ch <- obj
	}

	return ch
}
