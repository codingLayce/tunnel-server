package tunnel

import (
	"fmt"
	"sync"
)

type Broadcaster struct {
	name      string
	listeners map[string]func(msg string)

	mtx sync.Mutex
}

func NewBroadcaster(name string) *Broadcaster {
	return &Broadcaster{
		name:      name,
		listeners: make(map[string]func(msg string)),
	}
}

func (b *Broadcaster) PublishMessage(senderID, msg string) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	for listenerID, listener := range b.listeners {
		if listenerID == senderID {
			continue
		}
		listener(msg)
		// TODO: Maybe mange here net.Conn and timeout/retry mechanism
	}
}

func (b *Broadcaster) Name() string {
	return b.name
}

func (b *Broadcaster) RegisterListener(id string, callback func(tunnelName, msg string)) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	if _, ok := b.listeners[id]; ok {
		return fmt.Errorf("listener already exists")
	}
	b.listeners[id] = func(msg string) {
		callback(b.name, msg)
	}
	return nil
}

func (b *Broadcaster) UnregisterListener(id string) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	delete(b.listeners, id)
}
