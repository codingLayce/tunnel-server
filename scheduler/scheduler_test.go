package scheduler

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBroadcastTunnel(t *testing.T) {
	err := CreateBroadcastTunnel("MyTunnel")
	require.NoError(t, err)
	err = CreateBroadcastTunnel("SupperTunnel")
	require.NoError(t, err)
	err = CreateBroadcastTunnel("AnotherTunnel")
	require.NoError(t, err)
}

func TestCreateBroadcastTunnel_Concurrency(t *testing.T) {
	nbProcess := 1_000
	wg := sync.WaitGroup{}
	wg.Add(nbProcess)

	for i := 0; i < nbProcess; i++ {
		go func() {
			defer wg.Done()
			err := CreateBroadcastTunnel(fmt.Sprintf("Tunnel%d", i))
			require.NoError(t, err)
		}()
	}
	wg.Wait()
}

func TestCreateBroadcastTunnel_AlreadyExistsError(t *testing.T) {
	err := CreateBroadcastTunnel("Bidule")
	require.NoError(t, err)
	err = CreateBroadcastTunnel("Bidule")
	assert.EqualError(t, err, `tunnel named "Bidule" already exists`)
}
