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
	typeViews map[string][]*view.EphemeralView[any]
}

func NewBus() *Bus {
	return &Bus{
		keyViews:  map[string][]*view.EphemeralView[any]{},
		typeViews: map[string][]*view.EphemeralView[any]{},
	}
}

func (b *Bus) Announce(t string, obj any) {
	key := getObjKey(t, obj)

	b.mu.Lock()
	defer b.mu.Unlock()

	b.keyViews[key] = announce(obj, b.keyViews[key])
	b.typeViews[t] = announce(obj, b.typeViews[t])
}

func (b *Bus) Delete(t string, id string) {
	key := getKey(t, id)

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, ev := range b.keyViews[key] {
		ev.Close()
	}

	b.typeViews[t] = announce(id, b.typeViews[t])
}

func (b *Bus) SubscribeKey(t string, id string, ev *view.EphemeralView[any]) {
	key := getKey(t, id)

	b.mu.Lock()
	defer b.mu.Unlock()

	b.keyViews[key] = append(b.keyViews[key], ev)
}

func (b *Bus) SubscribeType(t string, ev *view.EphemeralView[any]) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.typeViews[t] = append(b.typeViews[t], ev)
}

func getObjKey(t string, obj any) string {
	return getKey(t, metadata.GetMetadata(obj).ID)
}

func getKey(t string, id string) string {
	return fmt.Sprintf("%s:%s", t, id)
}

func announce(obj any, views []*view.EphemeralView[any]) []*view.EphemeralView[any] {
	ret := []*view.EphemeralView[any]{}

	for _, ev := range views {
		err := ev.Update(obj)
		if err != nil {
			continue
		}

		ret = append(ret, ev)
	}

	return ret
}
