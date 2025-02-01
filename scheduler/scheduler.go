package scheduler

import (
	"fmt"
	"sync"

	"github.com/codingLayce/tunnel-server/tunnel"
)

var scheduler = &Scheduler{
	tunnels: make(map[string]tunnel.Tunnel),
}

// Scheduler is in charge of managing the Tunnels.
// It creates and deletes Tunnels.
// It receives the messages and dispatch them to the appropriate Tunnels.
type Scheduler struct {
	tunnels map[string]tunnel.Tunnel

	mtx sync.Mutex
}

func PublishMessage(clientID, tunnelName, msg string) error {
	tun, exists := scheduler.getTunnel(tunnelName)
	if !exists {
		return fmt.Errorf("unknown tunnel %q", tunnelName)
	}
	go tun.PublishMessage(clientID, msg) // TODO: Better management of this go routine
	return nil
}

func ListenTunnel(tunnelName, clientID string, callback func(tunnelName, msg string)) error {
	tun, exists := scheduler.getTunnel(tunnelName)
	if !exists {
		return fmt.Errorf("unknown tunnel %q", tunnelName)
	}

	err := tun.RegisterListener(clientID, callback)
	if err != nil {
		return fmt.Errorf("register listener: %w", err)
	}
	return nil
}

func CreateBroadcastTunnel(name string) error {
	broadcaster := tunnel.NewBroadcaster(name)
	if !scheduler.storeTunnel(broadcaster) {
		return fmt.Errorf("tunnel named %q already exists", name)
	}
	return nil
}

func StopAllListen(clientID string) {
	scheduler.stopAllListenOfClient(clientID)
}

func (s *Scheduler) storeTunnel(tunnel tunnel.Tunnel) bool {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, exists := s.tunnels[tunnel.Name()]; exists {
		return false
	}
	s.tunnels[tunnel.Name()] = tunnel
	return true
}

func (s *Scheduler) getTunnel(name string) (tunnel.Tunnel, bool) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	tun, ok := s.tunnels[name]
	return tun, ok
}

func (s *Scheduler) stopAllListenOfClient(clientID string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	for _, tun := range s.tunnels {
		// The client might not be listening that one but, it won't fail.
		// Maybe refactor later to only get the Tunnels the client is listening.
		tun.UnregisterListener(clientID)
	}
}
