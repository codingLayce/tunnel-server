package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/codingLayce/tunnel-server/tunnel"
	"github.com/codingLayce/tunnel.go/pdu"
	"github.com/codingLayce/tunnel.go/pdu/command"
)

// /!\ State is kept during all tests execution /!\

func TestCreateTunnel(t *testing.T) {
	srv, cli := setupServerAndClient(t)
	t.Cleanup(srv.Stop)
	t.Cleanup(cli.Stop)

	err := cli.Send(pdu.Marshal(command.NewCreateTunnel("Tunnel")))
	require.NoError(t, err)

	shouldReceiveAckBefore(t, cli, 100*time.Millisecond)
}

func TestCreateTunnel_TunnelAlreadyExists(t *testing.T) {
	err := tunnel.CreateBroadcast("MyTunnel")
	require.NoError(t, err)

	srv, cli := setupServerAndClient(t)
	t.Cleanup(srv.Stop)
	t.Cleanup(cli.Stop)

	err = cli.Send(pdu.Marshal(command.NewCreateTunnel("MyTunnel")))
	require.NoError(t, err)

	shouldReceiveNackBefore(t, cli, 100*time.Millisecond)
}
