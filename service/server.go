package service

import (
	"sync"

	"github.com/codingLayce/tunnel/tcp"
)

type Server struct {
	internal *tcp.Server

	clients map[string]*serverClient

	mtx sync.Mutex
}

func NewServer(addr string) *Server {
	server := &Server{
		clients: make(map[string]*serverClient),
	}
	server.internal = tcp.NewServer(&tcp.ServerOption{
		Addr:                 addr,
		OnConnectionReceived: server.connectionReceived,
		OnConnectionClosed:   server.connectionClosed,
		OnPayload:            server.payloadReceived,
	})

	return server
}

func (s *Server) connectionReceived(conn *tcp.Connection) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.clients[conn.ID] = newServerClient(conn.ID)
	s.clients[conn.ID].connected()
}

func (s *Server) connectionClosed(conn *tcp.Connection, timeout bool) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.clients[conn.ID].disconnected(timeout)
	delete(s.clients, conn.ID)
}
func (s *Server) payloadReceived(conn *tcp.Connection, payload []byte) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.clients[conn.ID].payloadReceived(payload)
}

func (s *Server) Start() error {
	return s.internal.Start()
}
func (s *Server) Stop() {
	s.internal.Stop()
}
func (s *Server) Done() <-chan struct{} {
	return s.internal.Done()
}
