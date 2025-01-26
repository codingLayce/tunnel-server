package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codingLayce/tunnel-server/scheduler"
	"github.com/codingLayce/tunnel.go/pdu"
	"github.com/codingLayce/tunnel.go/pdu/command"
)

// /!\ State is kept during all tests execution /!\

func TestListenTunnel(t *testing.T) {
	err := scheduler.CreateBroadcastTunnel("BTunnel")
	require.NoError(t, err)

	srv, cli := setupServerAndClient(t)
	defer srv.Stop()
	defer cli.Stop()

	err = cli.Send(pdu.Marshal(command.NewListenTunnel("BTunnel")))
	require.NoError(t, err)

	shouldReceiveAckBefore(t, cli, 100*time.Millisecond)

	err = scheduler.PublishMessage("SomeID", "BTunnel", "Un message de ouf")
	require.NoError(t, err)

	tunnelName, msg := shouldReceiveMessageBefore(t, cli, 100*time.Millisecond)
	assert.Equal(t, "BTunnel", tunnelName)
	assert.Equal(t, "Un message de ouf", msg)
}

func TestListenTunnel_TunnelDoesntExists(t *testing.T) {
	srv, cli := setupServerAndClient(t)
	defer srv.Stop()
	defer cli.Stop()

	err := cli.Send(pdu.Marshal(command.NewListenTunnel("UnknownTunnel")))
	require.NoError(t, err)

	shouldReceiveNackBefore(t, cli, 100*time.Millisecond)
}

func TestListenTunnel_DoubleListenForSameClient(t *testing.T) {
	err := scheduler.CreateBroadcastTunnel("TunnelDoubleListen")
	require.NoError(t, err)

	srv, cli := setupServerAndClient(t)
	defer srv.Stop()
	defer cli.Stop()

	err = cli.Send(pdu.Marshal(command.NewListenTunnel("TunnelDoubleListen")))
	require.NoError(t, err)

	shouldReceiveAckBefore(t, cli, 100*time.Millisecond)

	err = cli.Send(pdu.Marshal(command.NewListenTunnel("TunnelDoubleListen")))
	require.NoError(t, err)

	shouldReceiveNackBefore(t, cli, 100*time.Millisecond)
}
