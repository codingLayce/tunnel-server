package scheduler

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/codingLayce/tunnel-server/tunnel"
)

var scheduler = &Scheduler{
	tunnels: make(map[string]tunnel.Tunnel),
	logger:  slog.Default().With("entity", "SCHEDULER"),
}

// Scheduler is in charge of managing the Tunnels.
// It creates and deletes Tunnels.
// It receives the messages and dispatch them to the appropriate Tunnels.
type Scheduler struct {
	tunnels map[string]tunnel.Tunnel

	logger *slog.Logger

	mtx sync.Mutex
}

func PublishMessage(clientID, tunnelName, msg string) error {
	logger := scheduler.logger.With("tunnel_name", tunnelName)

	if !scheduler.tunnelExists(tunnelName) {
		logger.Warn("Unknown Tunnel")
		return fmt.Errorf("unknown tunnel %q", tunnelName)
	}
	tun := scheduler.getTunnel(tunnelName)
	tun.PublishMessage(clientID, msg)

	logger.Info("Message published to the Tunnel")
	return nil
}

func StopAllListen(clientID string) {
	scheduler.stopAllListenOfClient(clientID)
	scheduler.logger.Info("Client stopped listening to Tunnels", "client_id", clientID)
}

func ListenTunnel(name, clientID string, callback func(tunnelName, msg string)) error {
	logger := scheduler.logger.With("tunnel_name", name)
	logger.Debug("Listen Tunnel")

	if !scheduler.tunnelExists(name) {
		logger.Warn("Unknown Tunnel")
		return fmt.Errorf("unknown tunnel %q", name)
	}

	tun := scheduler.getTunnel(name)
	err := tun.RegisterListener(clientID, callback)
	if err != nil {
		logger.Warn("Cannot listen Tunnel", "error", err)
		return fmt.Errorf("register listener: %w", err)
	}

	return nil
}

func CreateBroadcastTunnel(name string) error {
	logger := scheduler.logger.With("tunnel_name", name)
	logger.Debug("Create a broadcast Tunnel")

	if scheduler.tunnelExists(name) {
		logger.Warn("A Tunnel with the same name already exists")
		return fmt.Errorf("tunnel named %q already exists", name)
	}

	broadcaster := tunnel.NewBroadcaster(name)
	scheduler.storeTunnel(broadcaster)

	logger.Info("Broadcast Tunnel created")

	return nil
}

func (s *Scheduler) storeTunnel(tunnel tunnel.Tunnel) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.tunnels[tunnel.Name()] = tunnel
}

func (s *Scheduler) getTunnel(name string) tunnel.Tunnel {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.tunnels[name]
}

func (s *Scheduler) tunnelExists(name string) bool {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	_, ok := s.tunnels[name]
	return ok
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
