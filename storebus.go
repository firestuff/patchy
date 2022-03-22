package storebus

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

func (sb *StoreBus) Write(obj Object) error {
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
	return sb.bus.Subscribe(obj)
}
