package helpers

import (
	"github.com/codingLayce/tunnel.go/pdu"
	"github.com/codingLayce/tunnel.go/pdu/command"
	"github.com/codingLayce/tunnel.go/tcp"
)

type ClientSpy struct {
	*tcp.Client
	commands chan command.Command
}

func NewClientSpy(addr string) *ClientSpy {
	client := &ClientSpy{
		commands: make(chan command.Command),
	}
	client.Client = tcp.NewClient(&tcp.ClientOption{
		Addr:      addr,
		OnPayload: client.onPayload,
	})
	return client
}

func (c *ClientSpy) onPayload(payload []byte) {
	cmd, _ := pdu.Unmarshal(payload) // Used only in tests and the server shouldn't send unparsable payloads
	c.commands <- cmd
}

func (c *ClientSpy) Commands() <-chan command.Command {
	return c.commands
}
