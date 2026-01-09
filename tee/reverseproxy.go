package tee

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type ReverseProxy struct {
	listener  net.Listener
	closeFunc CloseFunc
	serveFunc ServeFunc
}

func NewReverseProxy(
	ctx context.Context,
	platform Platform,
	addr string,
	targetAddr string,
	logger *slog.Logger,
) (*ReverseProxy, error) {
	dialContext, err := NewDialContext(platform)
	if err != nil {
		return nil, reverseProxyError("creating dialer", err)
	}
	return NewReverseProxyWithDialContext(ctx, dialContext, addr, targetAddr, logger)
}

func NewReverseProxyWithDialContext(
	ctx context.Context,
	dialContext DialContext,
	addr string,
	targetAddr string,
	logger *slog.Logger,
) (*ReverseProxy, error) {
	targetURL, err := url.Parse(targetAddr)
	if err != nil {
		return nil, reverseProxyError("parsing target URL", err)
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	reverseProxy.Transport = &http.Transport{DialContext: dialContext}

	// NOTE: Reverse Proxies are only ever run (1) outside a Nitro Enclave
	// or (2) within an SEV-SNP/TDX enclave. This means the reverse proxy will
	// always listen on a regular socket, which is why we use NoTEE here.
	listener, err := NewListener(ctx, NoTEE, NetworkTCP, addr)
	if err != nil {
		return nil, reverseProxyError("creating listener", err)
	}

	server := DefaultReverseProxyServer(reverseProxy, logger)
	closeFunc := func() error {
		if closeErr := listener.Close(); closeErr != nil {
			return closeErr
		}
		return server.Close()
	}
	serveFunc := func() error { return server.Serve(listener) }
	return &ReverseProxy{
		listener:  listener,
		closeFunc: closeFunc,
		serveFunc: serveFunc,
	}, nil
}

func NewReverseProxyTLS(
	ctx context.Context,
	platform Platform,
	addr string,
	targetAddr string,
	logger *slog.Logger,
) (*ReverseProxy, error) {
	dialContext, err := NewDialContext(platform)
	if err != nil {
		return nil, reverseProxyError("creating dialer", err)
	}
	return NewReverseProxyTLSWithDialContext(ctx, dialContext, addr, targetAddr, logger)
}

//nolint:contextcheck
func NewReverseProxyTLSWithDialContext(
	ctx context.Context,
	dialContext DialContext,
	addr string,
	targetAddr string,
	logger *slog.Logger,
) (*ReverseProxy, error) {
	// NOTE: Reverse Proxies are only ever run (1) outside a Nitro Enclave
	// or (2) within an SEV-SNP/TDX enclave. This means the reverse proxy will
	// always listen on a regular socket, which is why we use NoTEE here.
	listener, err := NewListener(ctx, NoTEE, NetworkTCP, addr)
	if err != nil {
		return nil, reverseProxyError("creating listener", err)
	}

	closeRevProxy := make(chan struct{})
	closeFunc := func() error {
		close(closeRevProxy)
		return listener.Close()
	}

	serveFunc := MakeReverseProxyTLSServeFunc(
		dialContext,
		listener,
		targetAddr,
		closeRevProxy,
		logger,
	)
	return &ReverseProxy{
		listener:  listener,
		closeFunc: closeFunc,
		serveFunc: serveFunc,
	}, nil
}

func (r *ReverseProxy) Addr() string { return r.listener.Addr().String() }
func (r *ReverseProxy) Close() error { return r.closeFunc() }
func (r *ReverseProxy) Serve() error { return r.serveFunc() }

func MakeReverseProxyTLSServeFunc(
	dialContext DialContext,
	listener net.Listener,
	targetAddr string,
	closeRevProxy chan struct{},
	logger *slog.Logger,
) ServeFunc {
	return func() error {
		for {
			select {
			case <-closeRevProxy:
				return nil
			default:
			}

			clientConn, err := listener.Accept()
			if err != nil {
				logger.Error("accepting connection", slog.String("error", err.Error()))
				continue
			}

			logger.Info("accepted connection", slog.String("addr", clientConn.RemoteAddr().String()))
			go proxyTLSConn(clientConn, dialContext, targetAddr, closeRevProxy, logger)
		}
	}
}

func proxyTLSConn(
	clientConn net.Conn,
	dialContext DialContext,
	targetAddr string,
	closeRevProxy chan struct{},
	logger *slog.Logger,
) {
	defer clientConn.Close()

	dialCtx, dialCancel := context.WithTimeout(context.Background(), DefaultConnTimeout)
	defer dialCancel()

	serverConn, err := dialContext(dialCtx, NetworkTCP, targetAddr)
	if err != nil {
		logger.Error("dialing target", slog.String("error", err.Error()))
		return
	}
	defer serverConn.Close()

	connCtx, connCancel := context.WithTimeout(context.Background(), DefaultConnTimeout)
	defer connCancel()

	connDone := make(chan error, NumConnDoneChannels)
	go func() {
		_, connErr := copyBuffer(serverConn, clientConn)
		connDone <- connErr
	}()
	go func() {
		_, connErr := copyBuffer(clientConn, serverConn)
		connDone <- connErr
	}()

	select {
	case connErr := <-connDone:
		if connErr != nil && !errors.Is(connErr, io.EOF) {
			logger.Error("conn error", slog.String("error", connErr.Error()))
		}
		return
	case <-closeRevProxy:
		logger.Info("proxy shutdown signal received, closing connection")
		return
	case <-connCtx.Done():
		logger.Error("proxy timeout", slog.String("error", connCtx.Err().Error()))
		return
	}
}

// copyBuffer copies from src to dst without using splice, avoiding kernel issues on SEV/TDX
func copyBuffer(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		nr, err := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if err != nil {
			if err == io.EOF {
				return written, nil
			}
			return written, err
		}
	}
}

func DefaultReverseProxyServer(handler http.Handler, logger *slog.Logger) *http.Server {
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
