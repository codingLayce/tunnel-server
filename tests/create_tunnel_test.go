package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codingLayce/tunnel-server/scheduler"
	"github.com/codingLayce/tunnel-server/server"
	"github.com/codingLayce/tunnel-server/tests/helpers"
	"github.com/codingLayce/tunnel.go/pdu"
	"github.com/codingLayce/tunnel.go/pdu/command"
)

// Those tests are meant to be end-to-end tests.

func TestServerStop(t *testing.T) {
	srv := server.NewServer(":0")
	err := srv.Start()
	require.NoError(t, err)

	go srv.Stop()

	select {
	case <-srv.Done():
	case <-time.After(100 * time.Millisecond):
		assert.FailNow(t, "Server should have stopped")
	}
}

func TestServerClient_ConnectionTimeout(t *testing.T) {
	// TODO: When ReadTimeout is configurable
	t.Skipf("When ReadTimeout will be configurable")
}

func TestServerClient_UnmarshablePayload(t *testing.T) {
	srv := server.NewServer(":0")
	err := srv.Start()
	require.NoError(t, err)
	defer srv.Stop()

	cli := helpers.NewClientSpy(srv.Addr())
	err = cli.Connect()
	require.NoError(t, err)
	defer cli.Stop()

	err = cli.Send([]byte("toto\n"))
	require.NoError(t, err)

	select {
	case <-cli.Commands():
		assert.FailNow(t, "No command should have been received")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestServerClient_UnsupportedCommand(t *testing.T) {
	srv := server.NewServer(":0")
	err := srv.Start()
	require.NoError(t, err)
	defer srv.Stop()

	cli := helpers.NewClientSpy(srv.Addr())
	err = cli.Connect()
	require.NoError(t, err)
	defer cli.Stop()

	err = cli.Send(pdu.Marshal(command.NewAck()))
	require.NoError(t, err)

	select {
	case <-cli.Commands():
		assert.FailNow(t, "No command should have been received")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestCreateTunnel(t *testing.T) {
	srv := server.NewServer(":0")
	err := srv.Start()
	require.NoError(t, err)
	defer srv.Stop()

	cli := helpers.NewClientSpy(srv.Addr())
	err = cli.Connect()
	require.NoError(t, err)
	defer cli.Stop()

	err = cli.Send(pdu.Marshal(command.NewCreateTunnel("Tunnel")))
	require.NoError(t, err)

	select {
	case cmd := <-cli.Commands():
		_, ok := cmd.(*command.Ack)
		assert.True(t, ok, "Command should be an ack")
	case <-time.After(100 * time.Millisecond):
		assert.FailNow(t, "Ack command should have been received")
	}
}

func TestCreateTunnel_TunnelAlreadyExists(t *testing.T) {
	err := scheduler.CreateBroadcastTunnel("MyTunnel")
	require.NoError(t, err)

	srv := server.NewServer(":0")
	err = srv.Start()
	require.NoError(t, err)
	defer srv.Stop()

	cli := helpers.NewClientSpy(srv.Addr())
	err = cli.Connect()
	require.NoError(t, err)
	defer cli.Stop()

	err = cli.Send(pdu.Marshal(command.NewCreateTunnel("MyTunnel")))
	require.NoError(t, err)

	select {
	case cmd := <-cli.Commands():
		_, ok := cmd.(*command.Nack)
		assert.True(t, ok, "Command should be a nack")
	case <-time.After(100 * time.Millisecond):
		assert.FailNow(t, "Nack command should have been received")
	}
}
