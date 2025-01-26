package server

import (
	"log/slog"

	"github.com/codingLayce/tunnel-server/scheduler"
	"github.com/codingLayce/tunnel.go/pdu"
	"github.com/codingLayce/tunnel.go/pdu/command"
	"github.com/codingLayce/tunnel.go/tcp"
)

type serverClient struct {
	conn   *tcp.Connection
	logger *slog.Logger
}

func newServerClient(conn *tcp.Connection) *serverClient {
	return &serverClient{
		conn:   conn,
		logger: slog.Default().With("entity", "CLIENT", "client", conn.ID),
	}
}

func (s *serverClient) payloadReceived(payload []byte) {
	s.logger.Debug("Received payload", "payload", string(payload))

	cmd, err := pdu.Unmarshal(payload)
	if err != nil {
		s.logger.Warn("Unparsable payload. Ignoring it", "error", err)
		return
	}

	// Contextual logger for the command's process
	logger := s.logger.With("command", cmd.Info(), "transaction_id", cmd.TransactionID())
	logger.Info("Command parsed")

	switch castedCMD := cmd.(type) {
	case *command.CreateTunnel:
		s.handleCreateTunnelCmd(logger, castedCMD)
	case *command.ListenTunnel:
		s.handleListenTunnel(logger, castedCMD)
	case *command.PublishMessage:
		s.handlePublishMessage(logger, castedCMD)
	default:
		logger.Warn("Unsupported command. Ignoring it")
	}
}

func (s *serverClient) handlePublishMessage(logger *slog.Logger, cmd *command.PublishMessage) {
	err := scheduler.PublishMessage(s.conn.ID, cmd.TunnelName, cmd.Message)
	if err != nil {
		logger.Warn("Cannot publish message", "error", err)
		s.nack(logger, cmd.TransactionID()) // TODO: Add reason to nack
		return
	}

	logger.Info("Message published to Tunnel", "tunnel_name", cmd.TunnelName)
	s.ack(logger, cmd.TransactionID())
}

func (s *serverClient) handleListenTunnel(logger *slog.Logger, cmd *command.ListenTunnel) {
	err := scheduler.ListenTunnel(cmd.Name, s.conn.ID, s.notifyMessageForTunnel)
	if err != nil {
		logger.Warn("Cannot listen Tunnel", "error", err)
		s.nack(logger, cmd.TransactionID()) // TODO: Add reason to nack
		return
	}
	logger.Info("Listen Tunnel")
	s.ack(logger, cmd.TransactionID())
}

func (s *serverClient) handleCreateTunnelCmd(logger *slog.Logger, cmd *command.CreateTunnel) {
	err := scheduler.CreateBroadcastTunnel(cmd.Name)
	if err != nil {
		logger.Warn("Cannot create broadcast Tunnel", "error", err)
		s.nack(logger, cmd.TransactionID()) // TODO: Add reason to nack
		return
	}
	logger.Info("Broadcast Tunnel created")
	s.ack(logger, cmd.TransactionID())
}

func (s *serverClient) ack(logger *slog.Logger, transactionID string) {
	payload := pdu.Marshal(command.NewAckWithTransactionID(transactionID))
	logger.Debug("Sending payload", "payload", payload)

	// TODO: Configure Write timeout
	_, err := s.conn.Write(payload)
	if err != nil {
		logger.Error("Cannot send ack", "error", err)
		return
	}
	logger.Info("Ack sent")
}

func (s *serverClient) nack(logger *slog.Logger, transactionID string) {
	payload := pdu.Marshal(command.NewNackWithTransactionID(transactionID))
	logger.Debug("Sending payload", "payload", payload)

	// TODO: Configure Write timeout
	_, err := s.conn.Write(payload)
	if err != nil {
		logger.Error("Cannot send nack", "error", err)
		return
	}
	logger.Info("Nack sent")
}

func (s *serverClient) connected() {
	s.logger.Info("Connected")
}

func (s *serverClient) disconnected(timeout bool) {
	s.logger.Info("Disconnecting...")
	scheduler.StopAllListen(s.conn.ID)
	if timeout {
		s.logger.Info("Timeout. Disconnected")
	} else {
		s.logger.Info("Disconnected")
	}
}

func (s *serverClient) notifyMessageForTunnel(tunnel, msg string) {
	logger := s.logger.With("tunnel_name", tunnel)
	logger.Info("Notifying message for Tunnel", "tunnel_name", tunnel)

	cmd := command.NewReceiveMessage(tunnel, msg)
	err := cmd.Validate()
	if err != nil {
		logger.Error("Cannot validate receive message command", "error", err)
		return
	}

	payload := pdu.Marshal(cmd)
	logger.Debug("Sending payload", "payload", payload)

	// TODO: Configure Write timeout
	_, err = s.conn.Write(payload)
	if err != nil {
		logger.Error("Cannot send message", "error", err)
		return
	}

	// TODO: Wait for ACK
}
