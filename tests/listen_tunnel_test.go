package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codingLayce/tunnel-server/server"
	"github.com/codingLayce/tunnel-server/tunnel"
	"github.com/codingLayce/tunnel.go/pdu"
	"github.com/codingLayce/tunnel.go/pdu/command"
	"github.com/codingLayce/tunnel.go/test-helper/mock"
)

// /!\ State is kept during all tests execution /!\

func TestListenTunnel(t *testing.T) {
	err := tunnel.CreateBroadcast("BTunnel")
	require.NoError(t, err)

	srv, cli := setupServerAndClient(t)
	t.Cleanup(srv.Stop)
	t.Cleanup(cli.Stop)

	err = cli.Send(pdu.Marshal(command.NewListenTunnel("BTunnel")))
	require.NoError(t, err)

	shouldReceiveAckBefore(t, cli, 100*time.Millisecond)

	go func() { // PublishMessage will be waiting for the client's ack. So it will block the ack if in the same goroutine.
		err = tunnel.PublishMessage("SomeID", "BTunnel", "Un message de ouf")
		require.NoError(t, err)
	}()

	tunnelName, msg := shouldReceiveMessageAndAckBefore(t, cli, 100*time.Millisecond)
	assert.Equal(t, "BTunnel", tunnelName)
	assert.Equal(t, "Un message de ouf", msg)
}

func TestListenTunnel_NackMessage(t *testing.T) {
	err := tunnel.CreateBroadcast("BTunnelNack")
	require.NoError(t, err)

	srv, cli := setupServerAndClient(t)
	t.Cleanup(srv.Stop)
	t.Cleanup(cli.Stop)

	err = cli.Send(pdu.Marshal(command.NewListenTunnel("BTunnelNack")))
	require.NoError(t, err)

	shouldReceiveAckBefore(t, cli, 100*time.Millisecond)

	go func() { // PublishMessage will be waiting for the client's ack. So it will block the ack if in the same goroutine.
		err = tunnel.PublishMessage("SomeID2", "BTunnelNack", "Un message de ouf")
		require.NoError(t, err)
	}()

	tunnelName, msg := shouldReceiveMessageAndNackBefore(t, cli, 100*time.Millisecond)
	assert.Equal(t, "BTunnelNack", tunnelName)
	assert.Equal(t, "Un message de ouf", msg)
}

func TestListenTunnel_AckTimeout(t *testing.T) {
	mock.Do(t, &server.MessageAckTimeout, time.Second)
	err := tunnel.CreateBroadcast("BTunnelTimeout")
	require.NoError(t, err)

	srv, cli := setupServerAndClient(t)
	t.Cleanup(srv.Stop)
	t.Cleanup(cli.Stop)

	err = cli.Send(pdu.Marshal(command.NewListenTunnel("BTunnelTimeout")))
	require.NoError(t, err)

	shouldReceiveAckBefore(t, cli, 100*time.Millisecond)

	go func() {
		err = tunnel.PublishMessage("SomeID3", "BTunnelTimeout", "Un message de ouf")
		require.NoError(t, err)
	}()

	// Wait for message and doesn't respond
	select {
	case cmd := <-cli.Commands():
		_, ok := cmd.(*command.ReceiveMessage)
		assert.True(t, ok, "Command should be a ReceiveMessage")
	case <-time.After(10 * time.Millisecond):
		assert.FailNow(t, "Nack command should have been received")
	}
}

func TestListenTunnel_TunnelDoesntExists(t *testing.T) {
	srv, cli := setupServerAndClient(t)
	t.Cleanup(srv.Stop)
	t.Cleanup(cli.Stop)

	err := cli.Send(pdu.Marshal(command.NewListenTunnel("UnknownTunnel")))
	require.NoError(t, err)

	shouldReceiveNackBefore(t, cli, 100*time.Millisecond)
}

func TestListenTunnel_DoubleListenForSameClient(t *testing.T) {
	err := tunnel.CreateBroadcast("TunnelDoubleListen")
	require.NoError(t, err)

	srv, cli := setupServerAndClient(t)
	t.Cleanup(srv.Stop)
	t.Cleanup(cli.Stop)

	err = cli.Send(pdu.Marshal(command.NewListenTunnel("TunnelDoubleListen")))
	require.NoError(t, err)

	shouldReceiveAckBefore(t, cli, 100*time.Millisecond)

	err = cli.Send(pdu.Marshal(command.NewListenTunnel("TunnelDoubleListen")))
	require.NoError(t, err)

	shouldReceiveAckBefore(t, cli, 100*time.Millisecond)
}
