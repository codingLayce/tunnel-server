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

func (s *Scheduler) storeTunnel(tunnel tunnel.Tunnel) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.tunnels[tunnel.Name()] = tunnel
}

func (s *Scheduler) tunnelExists(name string) bool {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	_, ok := s.tunnels[name]
	return ok
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
