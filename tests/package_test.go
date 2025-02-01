package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codingLayce/tunnel-server/server"
	"github.com/codingLayce/tunnel-server/tests/helpers"
	"github.com/codingLayce/tunnel.go/pdu"
	"github.com/codingLayce/tunnel.go/pdu/command"
)

func setupServer(t *testing.T) *server.Server {
	srv := server.NewServer(":0")
	err := srv.Start()
	require.NoError(t, err)
	return srv
}

func setupClient(t *testing.T, addr string) *helpers.ClientSpy {
	cli := helpers.NewClientSpy(addr)
	err := cli.Connect()
	require.NoError(t, err)
	return cli
}

func setupServerAndClient(t *testing.T) (*server.Server, *helpers.ClientSpy) {
	srv := setupServer(t)
	return srv, setupClient(t, srv.Addr())
}

func shouldReceiveAckBefore(t *testing.T, cli *helpers.ClientSpy, timeout time.Duration) {
	select {
	case cmd := <-cli.Commands():
		_, ok := cmd.(*command.Ack)
		assert.True(t, ok, "Command should be an ack")
	case <-time.After(timeout):
		assert.FailNow(t, "Ack command should have been received")
	}
}

func shouldReceiveNackBefore(t *testing.T, cli *helpers.ClientSpy, timeout time.Duration) {
	select {
	case cmd := <-cli.Commands():
		_, ok := cmd.(*command.Nack)
		assert.True(t, ok, "Command should be a nack")
	case <-time.After(timeout):
		assert.FailNow(t, "Nack command should have been received")
	}
}

func shouldReceiveMessageAndAckBefore(t *testing.T, cli *helpers.ClientSpy, timeout time.Duration) (tunnelName, message string) {
	select {
	case cmd := <-cli.Commands():
		receiveMessage, ok := cmd.(*command.ReceiveMessage)
		assert.True(t, ok, "Command should be a ReceiveMessage")
		err := cli.Send(pdu.Marshal(command.NewAckWithTransactionID(cmd.TransactionID())))
		require.NoError(t, err)
		return receiveMessage.TunnelName, receiveMessage.Message
	case <-time.After(timeout):
		assert.FailNow(t, "Nack command should have been received")
	}
	return "", ""
}

func shouldReceiveMessageAndNackBefore(t *testing.T, cli *helpers.ClientSpy, timeout time.Duration) (tunnelName, message string) {
	select {
	case cmd := <-cli.Commands():
		receiveMessage, ok := cmd.(*command.ReceiveMessage)
		assert.True(t, ok, "Command should be a ReceiveMessage")
		err := cli.Send(pdu.Marshal(command.NewNackWithTransactionID(cmd.TransactionID())))
		require.NoError(t, err)
		return receiveMessage.TunnelName, receiveMessage.Message
	case <-time.After(timeout):
		assert.FailNow(t, "Nack command should have been received")
	}
	return "", ""
}

func shouldNotReceiveCommandsBefore(t *testing.T, cli *helpers.ClientSpy, timeout time.Duration) {
	select {
	case <-cli.Commands():
		assert.FailNow(t, "No commands should have been received")
	case <-time.After(timeout):
	}
}
