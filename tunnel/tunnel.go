package tunnel

type Tunnel interface {
	Name() string
	RegisterListener(id string, callback func(tunnelName, msg string)) error
	UnregisterListener(id string)
	PublishMessage(senderID, msg string)
}
