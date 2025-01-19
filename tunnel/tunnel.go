package tunnel

type Tunnel interface {
	Name() string
	RegisterListener(id string) error // TODO: Currently the id is used but it will not be that
	UnregisterListener(id string)     // TODO: Currently the id is used but it will not be that
}
