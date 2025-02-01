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

func TestPublishMessage(t *testing.T) {
	tunnelName := "BTunnel_publish_message"
	err := tunnel.CreateBroadcast(tunnelName)
	require.NoError(t, err)

	srv, cli := setupServerAndClient(t)
	t.Cleanup(srv.Stop)
	t.Cleanup(cli.Stop)

	err = cli.Send(pdu.Marshal(command.NewPublishMessage(tunnelName, "Mon message de ouf")))
	require.NoError(t, err)

	shouldReceiveAckBefore(t, cli, 100*time.Millisecond)
}

func TestPublishMessage_UnknownTunnel(t *testing.T) {
	srv, cli := setupServerAndClient(t)
	t.Cleanup(srv.Stop)
	t.Cleanup(cli.Stop)

	err := cli.Send(pdu.Marshal(command.NewPublishMessage("BTunnel_publish_message_unknown", "Mon message de ouf")))
	require.NoError(t, err)

	shouldReceiveNackBefore(t, cli, 100*time.Millisecond)
}

func TestPublishMessage_MultipleListeners(t *testing.T) {
	srv := setupServer(t)
	t.Cleanup(srv.Stop)
	c1 := setupClient(t, srv.Addr())
	t.Cleanup(c1.Stop)
	c2 := setupClient(t, srv.Addr())
	t.Cleanup(c2.Stop)

	tunnelName := "BTunnel_publish_message_multiple_listeners"
	err := tunnel.CreateBroadcast(tunnelName)
	require.NoError(t, err)

	// Both clients listen
	err = c1.Send(pdu.Marshal(command.NewListenTunnel(tunnelName)))
	require.NoError(t, err)
	shouldReceiveAckBefore(t, c1, 100*time.Millisecond)
	err = c2.Send(pdu.Marshal(command.NewListenTunnel(tunnelName)))
	require.NoError(t, err)
	shouldReceiveAckBefore(t, c2, 100*time.Millisecond)

	err = tunnel.PublishMessage("nonExistingClientID", tunnelName, "Big message")
	require.NoError(t, err)

	// Both clients should receive message
	_, c1Msg := shouldReceiveMessageAndAckBefore(t, c1, 100*time.Millisecond)
	assert.Equal(t, "Big message", c1Msg)
	_, c2Msg := shouldReceiveMessageAndAckBefore(t, c2, 100*time.Millisecond)
	assert.Equal(t, "Big message", c2Msg)

	// c1 publish message
	go func() {
		err = c1.Send(pdu.Marshal(command.NewPublishMessage(tunnelName, "Another big message")))
		require.NoError(t, err)
		shouldReceiveAckBefore(t, c1, 100*time.Millisecond)
	}()

	// Only c2 should receive it
	_, c2Msg = shouldReceiveMessageAndAckBefore(t, c2, 100*time.Millisecond)
	assert.Equal(t, "Another big message", c2Msg)
	shouldNotReceiveCommandsBefore(t, c1, 100*time.Millisecond)
}
