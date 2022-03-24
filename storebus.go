package patchy

type StoreBus struct {
	store *Store
	bus   *Bus
}

func NewStoreBus(root string) *StoreBus {
	return &StoreBus{
		store: NewStore(root),
		bus:   NewBus(),
	}
}

func (sb *StoreBus) Write(t string, obj interface{}) error {
	err := sb.store.Write(t, obj)
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

func (sb *StoreBus) Subscribe(t string, obj interface{}) chan interface{} {
	return sb.bus.Subscribe(t, obj)
}
