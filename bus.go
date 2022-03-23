package storebus

import "sync"

type Bus struct {
	mu    sync.Mutex
	chans map[string][]chan Object
}

func NewBus() *Bus {
	return &Bus{
		chans: map[string][]chan Object{},
	}
}

func (b *Bus) Announce(obj Object) {
	key := ObjectKey(obj)

	b.mu.Lock()
	defer b.mu.Unlock()

	chans := b.chans[key]
	newChans := []chan Object{}

	for _, ch := range chans {
		select {
		case ch <- obj:
			newChans = append(newChans, ch)
		default:
			close(ch)
		}
	}

	if len(chans) != len(newChans) {
		b.chans[key] = newChans
	}
}

func (b *Bus) Delete(obj Object) {
	key := ObjectKey(obj)

	b.mu.Lock()
	defer b.mu.Unlock()

	chans := b.chans[key]
	for _, ch := range chans {
		close(ch)
	}

	delete(b.chans, key)
}

func (b *Bus) Subscribe(obj Object) chan Object {
	key := ObjectKey(obj)

	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Object, 100)

	b.chans[key] = append(b.chans[key], ch)

	return ch
}
