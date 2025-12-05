package tee

import (
	"fmt"
	"net"
	"net/http"

	"github.com/tahardi/bearclave"
)

type Server struct {
	listener net.Listener
	server   *http.Server
}

func NewServer(
	platform bearclave.Platform,
	network string,
	addr string,
	mux *http.ServeMux,
) (*Server, error) {
	listener, err := bearclave.NewListener(platform, network, addr)
	if err != nil {
		return nil, fmt.Errorf("creating listener: %w", err)
	}
	return NewServerWithListener(listener, mux)
}

func NewServerWithListener(
	listener net.Listener,
	mux *http.ServeMux,
) (*Server, error) {
	return &Server{
		listener: listener,
		server:   &http.Server{Handler: mux},
	}, nil
}

func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

func (s *Server) Close() error {
	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("closing listener: %w", err)
	}

	if err := s.server.Close(); err != nil {
		return fmt.Errorf("closing server: %w", err)
	}
	return nil
}

func (s *Server) Serve() error {
	return s.server.Serve(s.listener)
}
