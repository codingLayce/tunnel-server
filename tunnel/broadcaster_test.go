package tunnel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBroadcaster(t *testing.T) {
	broadcaster := NewBroadcaster("Tunnel")
	assert.Equal(t, "Tunnel", broadcaster.Name())
}
