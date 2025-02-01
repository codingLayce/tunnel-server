package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codingLayce/tunnel-server/tunnel"
	"github.com/codingLayce/tunnel.go/pdu"
	"github.com/codingLayce/tunnel.go/pdu/command"
)

// /!\ State is kept during all tests execution /!\

func TestNotifyMessage_InvalidMessage(t *testing.T) {
	tunnelName := "BTunnel_invalid_name"
	err := tunnel.CreateBroadcast(tunnelName)
	require.NoError(t, err)

	srv, cli := setupServerAndClient(t)
	t.Cleanup(srv.Stop)
	t.Cleanup(cli.Stop)

	err = cli.Send(pdu.Marshal(command.NewListenTunnel(tunnelName)))
	require.NoError(t, err)

	shouldReceiveAckBefore(t, cli, 100*time.Millisecond)

	// Publish invalid message (shouldn't be received by the client)
	err = tunnel.PublishMessage("ClientID", tunnelName, "Inv$alid m&ssage")
	require.NoError(t, err)

	select {
	case <-cli.Commands():
		assert.FailNow(t, "No command should have been received")
	case <-time.After(100 * time.Millisecond):
	}
}
