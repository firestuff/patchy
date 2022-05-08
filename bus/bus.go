package bus

import (
	"fmt"
	"sync"

	"github.com/firestuff/patchy/metadata"
	"github.com/firestuff/patchy/view"
)

type Bus struct {
	mu        sync.Mutex
	keyViews  map[string][]*view.EphemeralView[any]
	typeChans map[string][]chan any
	delChans  map[string][]chan string
}

func NewBus() *Bus {
	return &Bus{
		keyViews:  map[string][]*view.EphemeralView[any]{},
		typeChans: map[string][]chan any{},
		delChans:  map[string][]chan string{},
	}
}

func (b *Bus) Announce(t string, obj any) {
	key := getObjKey(t, obj)

	b.mu.Lock()
	defer b.mu.Unlock()

	keyViews := b.keyViews[key]
	newKeyViews := []*view.EphemeralView[any]{}

	for _, ev := range keyViews {
		err := ev.Update(obj)
		if err != nil {
			continue
		}

		newKeyViews = append(newKeyViews, ev)
	}

	if len(keyViews) != len(newKeyViews) {
		b.keyViews[key] = newKeyViews
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

	for _, ev := range b.keyViews[key] {
		ev.Close()
	}

	delete(b.keyViews, key)

	for _, ch := range b.delChans[t] {
		ch <- id
	}
}

func (b *Bus) SubscribeKey(t string, id string, ev *view.EphemeralView[any]) {
	key := getKey(t, id)

	b.mu.Lock()
	defer b.mu.Unlock()

	b.keyViews[key] = append(b.keyViews[key], ev)
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
	return getKey(t, metadata.GetMetadata(obj).ID)
}

func getKey(t string, id string) string {
	return fmt.Sprintf("%s:%s", t, id)
}
