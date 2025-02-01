package server

import (
	"github.com/codingLayce/tunnel-server/tunnel"
	"github.com/codingLayce/tunnel.go/common/maps"
	"github.com/codingLayce/tunnel.go/tcp"
)

type Server struct {
	internal *tcp.Server

	// TODO: Migrate to maps.SyncMap
	clients *maps.SyncMap[string, *serverClient]
}

func NewServer(addr string) *Server {
	srv := &Server{
		clients: maps.NewSyncMap[string, *serverClient](),
	}
	srv.internal = tcp.NewServer(&tcp.ServerOption{
		Addr:                 addr,
		OnConnectionReceived: srv.connectionReceived,
		OnConnectionClosed:   srv.connectionClosed,
		OnPayload:            srv.payloadReceived,
	})

	return srv
}

func (s *Server) connectionReceived(conn *tcp.Connection) {
	srvClient := newServerClient(conn)
	s.clients.Put(conn.ID, srvClient)
	srvClient.connected()
}

func (s *Server) connectionClosed(conn *tcp.Connection, timeout bool) {
	srvClient, exists := s.clients.Get(conn.ID)
	if !exists {
		return
	}
	srvClient.disconnected(timeout)
	s.clients.Delete(conn.ID)
}

func (s *Server) payloadReceived(conn *tcp.Connection, payload []byte) {
	srvClient, exists := s.clients.Get(conn.ID)
	if !exists {
		return
	}
	srvClient.payloadReceived(payload)
}

func (s *Server) Start() error {
	return s.internal.Start()
}

func (s *Server) Stop() {
	s.internal.Stop()
	tunnel.StopTunnels()
}

func (s *Server) Addr() string {
	return s.internal.Addr()
}

func (s *Server) Done() <-chan struct{} {
	return s.internal.Done()
}
