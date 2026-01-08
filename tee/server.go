package tee

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

const (
	Megabyte                 = 1 << 20
	DefaultReadHeaderTimeout = 10 * time.Second
	DefaultReadTimeout       = 15 * time.Second
	DefaultWriteTimeout      = 15 * time.Second
	DefaultIdleTimeout       = 60 * time.Second
	DefaultMaxHeaderBytes    = 1 * Megabyte // 1MB
)

type Server struct {
	listener net.Listener
	server   *http.Server
}

func NewServer(
	ctx context.Context,
	platform Platform,
	addr string,
	handler http.Handler,
) (*Server, error) {
	listener, err := NewListener(ctx, platform, Network, addr)
	if err != nil {
		return nil, serverError("creating listener", err)
	}
	return NewServerWithListener(listener, handler)
}

func NewServerWithListener(
	listener net.Listener,
	handler http.Handler,
) (*Server, error) {
	server := DefaultServer(handler)
	return &Server{
		listener: listener,
		server:   server,
	}, nil
}

func NewServerTLS(
	ctx context.Context,
	platform Platform,
	addr string,
	handler http.Handler,
	certProvider CertProvider,
) (*Server, error) {
	listener, err := NewListener(ctx, platform, Network, addr)
	if err != nil {
		return nil, serverError("creating listener", err)
	}
	return NewServerTLSWithListener(listener, handler, certProvider)
}

func NewServerTLSWithListener(
	listener net.Listener,
	handler http.Handler,
	certProvider CertProvider,
) (*Server, error) {
	tlsConfig := &tls.Config {
		GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			ctx, cancel := context.WithTimeout(context.Background(), DefaultConnTimeout)
			defer cancel()

			cert, err := certProvider.GetCert(ctx)
			if err != nil {
				return nil, err
			}
			return cert, nil
		},
	}

	server := DefaultServer(handler)
	server.TLSConfig = tlsConfig
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
		return serverError("closing listener", err)
	}

	if err := s.server.Close(); err != nil {
		return serverError("closing server", err)
	}
	return nil
}

// TODO: Consider using serve func instead
func (s *Server) Serve() error {
	if s.server.TLSConfig == nil {
		return s.server.Serve(s.listener)
	}

	tlsListener := tls.NewListener(s.listener, s.server.TLSConfig)
	s.listener = tlsListener
	return s.server.Serve(s.listener)
}

func DefaultServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		MaxHeaderBytes:    DefaultMaxHeaderBytes,
		IdleTimeout:       DefaultIdleTimeout,
		ReadHeaderTimeout: DefaultReadHeaderTimeout,
		ReadTimeout:       DefaultReadTimeout,
		WriteTimeout:      DefaultWriteTimeout,
	}
}
