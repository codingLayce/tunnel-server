package tunnel

import (
	"fmt"

	"github.com/codingLayce/tunnel.go/common/maps"
)

type (
	Tunnel interface {
		RegisterListener(listener Listener)
		UnregisterListener(id string)
		PublishMessage(msg Message)
		Stop()
	}
	Listener interface {
		ID() string
		NotifyMessage(tunnelName, message string)
	}
	Message struct {
		SenderID string
		Msg      string
	}
)

var tunnels = maps.NewSyncMap[string, Tunnel]()

func CreateBroadcast(tunnelName string) error {
	if tunnels.Has(tunnelName) {
		return fmt.Errorf("tunnel named %q already exists", tunnelName)
	}
	tunnels.Put(tunnelName, newBroadcaster(tunnelName))
	return nil
}

func Listen(tunnelName string, listener Listener) error {
	tunnel, exists := tunnels.Get(tunnelName)
	if !exists {
		return fmt.Errorf("unknown tunnel %q", tunnelName)
	}
	tunnel.RegisterListener(listener)
	return nil
}

func PublishMessage(senderID, tunnelName, msg string) error {
	tunnel, exists := tunnels.Get(tunnelName)
	if !exists {
		return fmt.Errorf("unknown tunnel %q", tunnelName)
	}
	// No problem "GOing" this method because on stop it will return the goroutine if blocked.
	go tunnel.PublishMessage(Message{
		SenderID: senderID,
		Msg:      msg,
	})
	return nil
}

func StopListen(clientID string) {
	tunnels.Foreach(func(_ string, tunnel Tunnel) {
		tunnel.UnregisterListener(clientID)
	})
}

func StopTunnels() {
	tunnels.Foreach(func(_ string, tunnel Tunnel) {
		tunnel.Stop()
	})
}
