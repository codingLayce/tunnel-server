package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codingLayce/tunnel-server/server"
	"github.com/codingLayce/tunnel-server/tests/helpers"
	"github.com/codingLayce/tunnel.go/pdu/command"
)

func setupServerAndClient(t *testing.T) (*server.Server, *helpers.ClientSpy) {
	srv := server.NewServer(":0")
	err := srv.Start()
	require.NoError(t, err)

	cli := helpers.NewClientSpy(srv.Addr())
	err = cli.Connect()
	require.NoError(t, err)
	return srv, cli
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
