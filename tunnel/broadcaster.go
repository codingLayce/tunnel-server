package tunnel

import (
	"context"
	"sync"

	"github.com/codingLayce/tunnel.go/common/maps"
)

type Broadcaster struct {
	name      string
	listeners *maps.SyncMap[string, Listener]
	messages  chan Message

	ctx    context.Context
	stopFn context.CancelFunc
	wg     sync.WaitGroup
}

func newBroadcaster(name string) *Broadcaster {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Broadcaster{
		name:      name,
		listeners: maps.NewSyncMap[string, Listener](),
		messages:  make(chan Message),
		ctx:       ctx,
		stopFn:    cancel,
	}
	go b.start()
	return b
}

func (b *Broadcaster) RegisterListener(listener Listener) {
	b.listeners.Put(listener.ID(), listener)
}

func (b *Broadcaster) UnregisterListener(id string) {
	b.listeners.Delete(id)
}

func (b *Broadcaster) PublishMessage(msg Message) {
	select {
	case b.messages <- msg:
	case <-b.ctx.Done():
	}
}

func (b *Broadcaster) start() {
	b.wg.Add(1)
	defer b.wg.Done()

	for {
		select {
		case msg := <-b.messages:
			b.listeners.Foreach(func(id string, listener Listener) {
				if msg.SenderID == id {
					return
				}
				// TODO: goroutine for each listener
				listener.NotifyMessage(b.name, msg.Msg)
			})
		case <-b.ctx.Done():
			return
		}
	}
}

func (b *Broadcaster) Stop() {
	b.stopFn()
	b.wg.Wait()
}
