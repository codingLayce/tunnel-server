package server

import (
	"sync"

	"github.com/codingLayce/tunnel.go/tcp"
)

type Server struct {
	internal *tcp.Server

	clients map[string]*serverClient

	mtx sync.Mutex
}

func NewServer(addr string) *Server {
	srv := &Server{
		clients: make(map[string]*serverClient),
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
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.clients[conn.ID] = newServerClient(conn)
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
