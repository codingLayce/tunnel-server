package tunnel

import (
	"fmt"
	"sync"
)

type Broadcaster struct {
	name      string
	listeners map[string]struct{}

	mtx sync.Mutex
}

func NewBroadcaster(name string) *Broadcaster {
	return &Broadcaster{
		name:      name,
		listeners: make(map[string]struct{}),
	}
}

func (b *Broadcaster) Name() string {
	return b.name
}

func (b *Broadcaster) RegisterListener(id string) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	if _, ok := b.listeners[id]; ok {
		return fmt.Errorf("listener already exists")
	}
	b.listeners[id] = struct{}{}
	return nil
}

func (b *Broadcaster) UnregisterListener(id string) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	delete(b.listeners, id)
}
