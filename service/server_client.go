package service

import "log/slog"

type serverClient struct {
	logger *slog.Logger
}

func newServerClient(id string) *serverClient {
	return &serverClient{
		logger: slog.Default().With("entity", "CLIENT", "client", id),
	}
}

func (s *serverClient) connected() {
	if s == nil { // Edge case but prevent npe
		return
	}
	s.logger.Info("Connected")
}
func (s *serverClient) disconnected(timeout bool) {
	if s == nil { // Edge case but prevent npe
		return
	}
	if timeout {
		s.logger.Info("Timeout. Disconnected")
	} else {
		s.logger.Info("Disconnected")
	}
}

func (s *serverClient) payloadReceived(payload []byte) {
	if s == nil { // Edge case but prevent npe
		return
	}
	s.logger.Info("Received payload", "payload", string(payload))
}
