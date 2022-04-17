package bus

import "fmt"
import "sync"

import "github.com/firestuff/patchy/metadata"

type Bus struct {
	mu        sync.Mutex
	keyChans  map[string][]chan any
	typeChans map[string][]chan any
	delChans  map[string][]chan string
}

func NewBus() *Bus {
	return &Bus{
		keyChans:  map[string][]chan any{},
		typeChans: map[string][]chan any{},
		delChans:  map[string][]chan string{},
	}
}

func (b *Bus) Announce(t string, obj any) {
	key := getObjKey(t, obj)

	b.mu.Lock()
	defer b.mu.Unlock()

	keyChans := b.keyChans[key]
	newKeyChans := []chan any{}

	for _, ch := range keyChans {
		select {
		case ch <- obj:
			newKeyChans = append(newKeyChans, ch)
		default:
			close(ch)
		}
	}

	if len(keyChans) != len(newKeyChans) {
		b.keyChans[key] = newKeyChans
	}

	typeChans := b.typeChans[t]
	newTypeChans := []chan any{}

	for _, ch := range typeChans {
		select {
		case ch <- obj:
			newTypeChans = append(newTypeChans, ch)
		default:
			close(ch)
		}
	}

	if len(typeChans) != len(newTypeChans) {
		b.typeChans[key] = newTypeChans
	}
}

func (b *Bus) Delete(t string, id string) {
	key := getKey(t, id)

	b.mu.Lock()
	defer b.mu.Unlock()

	keyChans := b.keyChans[key]
	for _, ch := range keyChans {
		close(ch)
	}

	delete(b.keyChans, key)

	delChans := b.delChans[t]
	for _, ch := range delChans {
		ch <- id
	}
}

func (b *Bus) SubscribeKey(t string, id string) chan any {
	key := getKey(t, id)

	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan any, 100)

	b.keyChans[key] = append(b.keyChans[key], ch)

	return ch
}

func (b *Bus) SubscribeType(t string) (chan any, chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	typeChan := make(chan any, 100)
	b.typeChans[t] = append(b.typeChans[t], typeChan)

	delChan := make(chan string, 100)
	b.delChans[t] = append(b.delChans[t], delChan)

	return typeChan, delChan
}

func getObjKey(t string, obj any) string {
	return getKey(t, metadata.GetMetadata(obj).Id)
}

func getKey(t string, id string) string {
	return fmt.Sprintf("%s:%s", t, id)
}
