package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codingLayce/tunnel-server/server"
	"github.com/codingLayce/tunnel.go/pdu"
	"github.com/codingLayce/tunnel.go/pdu/command"
)

// /!\ State is kept during all tests execution /!\

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

func TestServerClient_UnmarshablePayload(t *testing.T) {
	srv, cli := setupServerAndClient(t)
	t.Cleanup(srv.Stop)
	t.Cleanup(cli.Stop)

	err := cli.Send([]byte("toto\n"))
	require.NoError(t, err)

	select {
	case <-cli.Commands():
		assert.FailNow(t, "No command should have been received")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestServerClient_UnsupportedCommand(t *testing.T) {
	srv, cli := setupServerAndClient(t)
	t.Cleanup(srv.Stop)
	t.Cleanup(cli.Stop)

	err := cli.Send(pdu.Marshal(command.NewReceiveMessage("tunnel", "msg")))
	require.NoError(t, err)

	select {
	case <-cli.Commands():
		assert.FailNow(t, "No command should have been received")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestServerClient_AckWithoutWaiter(t *testing.T) {
	srv, cli := setupServerAndClient(t)
	t.Cleanup(srv.Stop)
	t.Cleanup(cli.Stop)

	err := cli.Send(pdu.Marshal(command.NewAck()))
	require.NoError(t, err)

	select {
	case <-cli.Commands():
		assert.FailNow(t, "No command should have been received")
	case <-time.After(100 * time.Millisecond):
	}
}
