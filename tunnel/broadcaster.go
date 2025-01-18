package tunnel

type Broadcaster struct {
	name string
}

func NewBroadcaster(name string) *Broadcaster {
	return &Broadcaster{name: name}
}

func (b *Broadcaster) Name() string {
	return b.name
}
