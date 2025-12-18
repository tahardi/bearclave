package tee

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/tahardi/bearclave"
)

const (
	Megabyte = 1 << 20
	DefaultReadHeaderTimeout = 10 * time.Second
	DefaultReadTimeout       = 15 * time.Second
	DefaultWriteTimeout      = 15 * time.Second
	DefaultIdleTimeout       = 60 * time.Second
	DefaultMaxHeaderBytes    = 1 * Megabyte // 1MB
)

type ServerOption func(server *http.Server)

func WithServerErrorLog(logger *log.Logger) ServerOption {
	return func(server *http.Server) {
		server.ErrorLog = logger
	}
}

func WithServerMaxHeaderBytes(bytes int) ServerOption {
	return func(server *http.Server) {
		server.MaxHeaderBytes = bytes
	}
}

func WithServerIdleTimeout(timeout time.Duration) ServerOption {
	return func(server *http.Server) {
		server.IdleTimeout = timeout
	}
}

func WithServerReadHeaderTimeout(timeout time.Duration) ServerOption {
	return func(server *http.Server) {
		server.ReadHeaderTimeout = timeout
	}
}

func WithServerReadTimeout(timeout time.Duration) ServerOption {
	return func(server *http.Server) {
		server.ReadTimeout = timeout
	}
}

func WithServerWriteTimeout(timeout time.Duration) ServerOption {
	return func(server *http.Server) {
		server.WriteTimeout = timeout
	}
}

type Server struct {
	listener net.Listener
	server   *http.Server
}

func NewServer(
	ctx context.Context,
	platform bearclave.Platform,
	network string,
	addr string,
	handler http.Handler,
	opts ...ServerOption,
) (*Server, error) {
	listener, err := bearclave.NewListener(ctx, platform, network, addr)
	if err != nil {
		return nil, fmt.Errorf("creating listener: %w", err)
	}
	return NewServerWithListener(listener, handler, opts...)
}

func NewServerWithListener(
	listener net.Listener,
	handler http.Handler,
	opts ...ServerOption,
) (*Server, error) {
	server := &http.Server{
		Handler:           handler,
		MaxHeaderBytes:    DefaultMaxHeaderBytes,
		IdleTimeout:       DefaultIdleTimeout,
		ReadHeaderTimeout: DefaultReadHeaderTimeout,
		ReadTimeout:       DefaultReadTimeout,
		WriteTimeout:      DefaultWriteTimeout,
	}
	for _, opt := range opts {
		opt(server)
	}

	return &Server{
		listener: listener,
		server:   server,
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

func (s *Server) Handler() http.Handler {
	return s.server.Handler
}

func (s *Server) ListenAndServe() error {
	return s.server.Serve(s.listener)
}
