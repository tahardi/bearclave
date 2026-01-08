package tee

import (
	"context"
	"log/slog"
	"net"
	"net/http"
)

type Proxy struct {
	listener net.Listener
	server   *http.Server
}

func NewProxy(
	ctx context.Context,
	platform Platform,
	addr string,
	client *http.Client,
	logger *slog.Logger,
) (*Proxy, error) {
	listener, err := NewListener(ctx, platform, NetworkTCP, addr)
	if err != nil {
		return nil, err
	}
	return NewProxyWithListener(client, logger, listener)
}

func NewProxyWithListener(
	client *http.Client,
	logger *slog.Logger,
	listener net.Listener,
) (*Proxy, error) {
	handler := MakeProxyHandler(client, logger, DefaultProxyTimeout)
	server := DefaultProxyServer(handler, logger)
	return &Proxy{
		server:   server,
		listener: listener,
	}, nil
}

func NewProxyTLS(
	ctx context.Context,
	platform Platform,
	addr string,
	logger *slog.Logger,
) (*Proxy, error) {
	listener, err := NewListener(ctx, platform, NetworkTCP, addr)
	if err != nil {
		return nil, err
	}
	return NewProxyTLSWithListener(logger, listener)
}

func NewProxyTLSWithListener(
	logger *slog.Logger,
	listener net.Listener,
) (*Proxy, error) {
	handler := MakeProxyTLSHandler(logger, DefaultProxyTimeout)
	server := DefaultProxyServer(handler, logger)
	return &Proxy{
		server:   server,
		listener: listener,
	}, nil
}

func (p *Proxy) Addr() string {
	return p.listener.Addr().String()
}

func (p *Proxy) Close() error {
	err := p.listener.Close()
	if err != nil {
		return err
	}

	err = p.server.Close()
	if err != nil {
		return err
	}
	return nil
}

func (p *Proxy) Serve() error {
	return p.server.Serve(p.listener)
}

func DefaultProxyServer(handler http.Handler, logger *slog.Logger) *http.Server {
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
