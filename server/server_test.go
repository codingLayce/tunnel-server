package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codingLayce/tunnel.go/pdu"
	"github.com/codingLayce/tunnel.go/pdu/command"
	"github.com/codingLayce/tunnel.go/tcp"
)

// The tests of the package are closed to end-to-end tests.
// Real tcp.Server and tcp.Client are exchanging payloads.
// The scheduler isn't mocked so it acts like when running.

func TestServer_StartAndStop(t *testing.T) {
	srv := NewServer(":0")
	err := srv.Start()
	require.NoError(t, err)

	go func() {
		srv.Stop()
	}()

	select {
	case <-srv.Done():
	case <-time.After(100 * time.Millisecond):
		assert.FailNow(t, "Server should have been stopped")
	}
}

func TestServer_ClientConnect(t *testing.T) {
	srv := NewServer(":0")
	err := srv.Start()
	require.NoError(t, err)
	defer srv.Stop()

	cli := tcp.NewClient(&tcp.ClientOption{
		Addr: srv.internal.Addr(),
	})
	err = cli.Connect()
	require.NoError(t, err)
	defer cli.Stop()
}

func TestServer_Handle_CreateTunnelCommand(t *testing.T) {
	srv := NewServer(":0")
	err := srv.Start()
	require.NoError(t, err)
	defer srv.Stop()

	receivedPayload := make(chan []byte)
	cli := tcp.NewClient(&tcp.ClientOption{
		Addr: srv.internal.Addr(),
		OnPayload: func(payload []byte) {
			receivedPayload <- payload
		},
	})
	err = cli.Connect()
	require.NoError(t, err)
	defer cli.Stop()

	err = cli.Send(pdu.Marshal(command.NewCreateTunnel("Bidule")))
	require.NoError(t, err)

	select {
	case payload := <-receivedPayload:
		assert.NotEmpty(t, payload)
	case <-time.After(100 * time.Millisecond):
		assert.FailNow(t, "A payload should have been received")
	}
}
