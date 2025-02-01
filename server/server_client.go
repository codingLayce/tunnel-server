package server

import (
	"log/slog"
	"time"

	"github.com/codingLayce/tunnel-server/scheduler"
	"github.com/codingLayce/tunnel.go/common/maps"
	"github.com/codingLayce/tunnel.go/pdu"
	"github.com/codingLayce/tunnel.go/pdu/command"
	"github.com/codingLayce/tunnel.go/tcp"
)

var MessageAckTimeout = 10 * time.Second

type serverClient struct {
	conn *tcp.Connection

	// ackWaiters stores channels waiting for an acknowledgement.
	// Writes true when ack, false otherwise.
	ackWaiters *maps.SyncMap[string, chan bool]

	close chan struct{}

	logger *slog.Logger
}

func newServerClient(conn *tcp.Connection) *serverClient {
	return &serverClient{
		conn:       conn,
		ackWaiters: maps.NewSyncMap[string, chan bool](),
		close:      make(chan struct{}),
		logger:     slog.Default().With("client", conn.ID),
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
	logger := s.logger.With("transaction_id", cmd.TransactionID(), "command", cmd.Info())
	logger.Debug("Command parsed")

	switch castedCMD := cmd.(type) {
	case *command.CreateTunnel:
		s.handleCreateTunnel(logger, castedCMD)
	case *command.ListenTunnel:
		s.handleListenTunnel(logger, castedCMD)
	case *command.PublishMessage:
		s.handlePublishMessage(logger, castedCMD)
	case *command.Ack:
		s.handleAcknowledgement(logger, castedCMD.TransactionID(), true)
	case *command.Nack:
		s.handleAcknowledgement(logger, castedCMD.TransactionID(), false)
	default:
		logger.Warn("Unsupported command. Ignoring it")
	}
}

func (s *serverClient) handleAcknowledgement(logger *slog.Logger, transactionID string, isAck bool) {
	waiter, exists := s.ackWaiters.Get(transactionID)
	if !exists {
		logger.Warn("No waiter for the given acknowledgement. Ignoring it.")
		return
	}
	select { // Try to push to waiter while connection isn't closed.
	case waiter <- isAck:
	case <-s.close:
	}
}

func (s *serverClient) handlePublishMessage(logger *slog.Logger, cmd *command.PublishMessage) {
	err := scheduler.PublishMessage(s.conn.ID, cmd.TunnelName, cmd.Message)
	if err != nil {
		logger.Warn("Cannot publish message", "error", err)
		s.nack(logger, cmd.TransactionID()) // TODO: Add reason to nack
		return
	}
	s.ack(logger, cmd.TransactionID())
	logger.Info("Message published to Tunnel", "tunnel_name", cmd.TunnelName)
}

func (s *serverClient) handleListenTunnel(logger *slog.Logger, cmd *command.ListenTunnel) {
	err := scheduler.ListenTunnel(cmd.Name, s.conn.ID, s.notifyMessageForTunnel)
	if err != nil {
		logger.Warn("Cannot listen Tunnel", "error", err)
		s.nack(logger, cmd.TransactionID()) // TODO: Add reason to nack
		return
	}
	s.ack(logger, cmd.TransactionID())
	logger.Info("Listen Tunnel")
}

func (s *serverClient) handleCreateTunnel(logger *slog.Logger, cmd *command.CreateTunnel) {
	err := scheduler.CreateBroadcastTunnel(cmd.Name)
	if err != nil {
		logger.Warn("Cannot create broadcast Tunnel", "error", err)
		s.nack(logger, cmd.TransactionID()) // TODO: Add reason to nack
		return
	}
	s.ack(logger, cmd.TransactionID())
	logger.Info("Broadcast Tunnel created")
}

func (s *serverClient) notifyMessageForTunnel(tunnelName, msg string) {
	cmd := command.NewReceiveMessage(tunnelName, msg)
	logger := s.logger.With("transaction_id", cmd.TransactionID(), "command", cmd.Info())

	err := cmd.Validate()
	if err != nil {
		logger.Error("Cannot validate receive message command", "error", err)
		return
	}

	payload := pdu.Marshal(cmd)
	logger.Debug("Sending payload", "payload", payload)

	// Register waiter before sending payload in case of really fast networking (mainly for tests to be honest)
	ackCh := make(chan bool)
	s.ackWaiters.Put(cmd.TransactionID(), ackCh)
	defer s.ackWaiters.Delete(cmd.TransactionID())

	// TODO: Configure Write timeout
	_, err = s.conn.Write(payload)
	if err != nil {
		logger.Error("Cannot send message", "error", err)
		return
	}

	logger.Info("Message sent")

	select {
	case isAck := <-ackCh:
		if isAck {
			logger.Info("Message acked by client")
		} else {
			logger.Info("Message nacked by client")
		}
	case <-time.After(MessageAckTimeout):
		logger.Warn("Timeout waiting for client ack. Discard message")
	}
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
	logger.Debug("Ack sent")
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
	close(s.close)
	scheduler.StopAllListen(s.conn.ID)
	if timeout {
		s.logger.Info("Timeout. Disconnected")
	} else {
		s.logger.Info("Disconnected")
	}
}
