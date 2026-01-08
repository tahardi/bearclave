package tee

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"net/http"
	"time"
)

const (
	Megabyte                 = 1 << 20
	DefaultConnTimeout       = 15 * time.Second
	DefaultReadHeaderTimeout = 10 * time.Second
	DefaultReadTimeout       = 15 * time.Second
	DefaultWriteTimeout      = 15 * time.Second
	DefaultIdleTimeout       = 60 * time.Second
	DefaultMaxHeaderBytes    = 1 * Megabyte // 1MB
	NetworkTCP               = "tcp"
)

type CloseFunc func() error
type ServeFunc func() error
type Server struct {
	listener  net.Listener
	closeFunc CloseFunc
	serveFunc ServeFunc
}

func NewServer(
	ctx context.Context,
	platform Platform,
	addr string,
	handler http.Handler,
	logger *slog.Logger,
) (*Server, error) {
	listener, err := NewListener(ctx, platform, NetworkTCP, addr)
	if err != nil {
		return nil, serverError("creating listener", err)
	}
	return NewServerWithListener(listener, handler, logger)
}

func NewServerWithListener(
	listener net.Listener,
	handler http.Handler,
	logger *slog.Logger,
) (*Server, error) {
	server := DefaultServer(handler, logger)
	closeFunc := func() error {
		if closeErr := listener.Close(); closeErr != nil {
			return closeErr
		}
		return server.Close()
	}
	serverFunc := func() error { return server.Serve(listener) }
	return &Server{
		listener:  listener,
		closeFunc: closeFunc,
		serveFunc: serverFunc,
	}, nil
}

func NewServerTLS(
	ctx context.Context,
	platform Platform,
	addr string,
	handler http.Handler,
	certProvider CertProvider,
	logger *slog.Logger,
) (*Server, error) {
	listener, err := NewListener(ctx, platform, NetworkTCP, addr)
	if err != nil {
		return nil, serverError("creating listener", err)
	}
	return NewServerTLSWithListener(listener, handler, certProvider, logger)
}

func NewServerTLSWithListener(
	listener net.Listener,
	handler http.Handler,
	certProvider CertProvider,
	logger *slog.Logger,
) (*Server, error) {
	tlsConfig := &tls.Config{
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

	server := DefaultServer(handler, logger)
	closeFunc := func() error {
		if closeErr := listener.Close(); closeErr != nil {
			return closeErr
		}
		return server.Close()
	}
	serveFunc := func() error {
		tlsListener := tls.NewListener(listener, tlsConfig)
		return server.Serve(tlsListener)
	}
	return &Server{
		listener:  listener,
		closeFunc: closeFunc,
		serveFunc: serveFunc,
	}, nil
}

func (s *Server) Addr() string { return s.listener.Addr().String() }
func (s *Server) Close() error { return s.closeFunc() }
func (s *Server) Serve() error { return s.serveFunc() }

func DefaultServer(handler http.Handler, logger *slog.Logger) *http.Server {
	return &http.Server{
		Handler:           handler,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
		MaxHeaderBytes:    DefaultMaxHeaderBytes,
		IdleTimeout:       DefaultIdleTimeout,
		ReadHeaderTimeout: DefaultReadHeaderTimeout,
		ReadTimeout:       DefaultReadTimeout,
		WriteTimeout:      DefaultWriteTimeout,
	}
}
