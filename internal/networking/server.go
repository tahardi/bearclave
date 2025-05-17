package networking

import (
	"fmt"
	"net"
	"net/http"

	"github.com/mdlayher/vsock"

	"github.com/tahardi/bearclave/internal/setup"
)

type Server struct {
	listener net.Listener
	server   *http.Server
}

func NewServer(
	platform setup.Platform,
	port int,
	mux *http.ServeMux,
) (*Server, error) {
	switch platform {
	case setup.Nitro:
		return NewVSocketServer(port, mux)
	case setup.SEV:
		return NewSocketServer(port, mux)
	case setup.TDX:
		return NewSocketServer(port, mux)
	case setup.NoTEE:
		return NewSocketServer(port, mux)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}

func NewSocketServer(
	port int,
	mux *http.ServeMux,
) (*Server, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf(
			"setting up TCP socket listener on %s: %w",
			addr,
			err,
		)
	}
	return &Server{
		listener: listener,
		server:   &http.Server{Handler: mux},
	}, nil
}

func NewVSocketServer(
	port int,
	mux *http.ServeMux,
) (*Server, error) {
	listener, err := vsock.Listen(uint32(port), nil)
	if err != nil {
		return nil, fmt.Errorf(
			"setting up TCP vsocket listener on %d: %w",
			port,
			err,
		)
	}
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
