package bus

import "sync"

import "github.com/firestuff/patchy/metadata"

type Bus struct {
	mu    sync.Mutex
	chans map[string][]chan any
}

func NewBus() *Bus {
	return &Bus{
		chans: map[string][]chan any{},
	}
}

func (b *Bus) Announce(t string, obj any) {
	key := metadata.GetMetadata(obj).GetKey(t)

	b.mu.Lock()
	defer b.mu.Unlock()

	chans := b.chans[key]
	newChans := []chan any{}

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

func (b *Bus) Delete(t string, obj any) {
	key := metadata.GetMetadata(obj).GetKey(t)

	b.mu.Lock()
	defer b.mu.Unlock()

	chans := b.chans[key]
	for _, ch := range chans {
		close(ch)
	}

	delete(b.chans, key)
}

func (b *Bus) Subscribe(t string, obj any) chan any {
	key := metadata.GetMetadata(obj).GetKey(t)

	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan any, 100)

	b.chans[key] = append(b.chans[key], ch)

	return ch
}
