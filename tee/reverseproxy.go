package tee

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type CloseFunc func() error
type ServeFunc func() error
type ReverseProxy struct {
	listener  net.Listener
	closeFunc CloseFunc
	serveFunc ServeFunc
}

// TODO: Do I want route?
func NewReverseProxy(
	ctx context.Context,
	platform Platform,
	addr string,
	targetAddr string,
	route string,
) (*ReverseProxy, error) {
	dialContext, err := NewDialContext(platform)
	if err != nil {
		return nil, reverseProxyError("creating dialer", err)
	}
	return NewReverseProxyWithDialContext(ctx, dialContext, addr, targetAddr, route)
}

func NewReverseProxyWithDialContext(
	ctx context.Context,
	dialContext DialContext,
	addr string,
	targetAddr string,
	route string,
) (*ReverseProxy, error) {
	targetURL, err := url.Parse(targetAddr)
	if err != nil {
		return nil, reverseProxyError("parsing target URL", err)
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	reverseProxy.Transport = &http.Transport{DialContext: dialContext}

	// Customize the Director to strip the path prefix
	originalDirector := reverseProxy.Director
	reverseProxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, route)
	}

	// TODO: Explicitly use net to create listener to reduce chance of misconfig
	// NOTE: We assume that reverse proxies are only ever run (1) outside a
	// Nitro Enclave or (2) within an SEV-SNP/TDX enclave. For Nitro, the
	// DialContext must use vsock to send to the Enclave, but it must use a
	// normal socket to listen for inbound connections from remote clients.
	// This is why we enforce a normal socket listener here. Even though a
	// reverse proxy runs within the trusted zone for SEV-SNP and TDX, they
	// both use normal sockets, so using a socket listener here is fine.
	listener, err := NewListener(ctx, NoTEE, Network, addr)
	if err != nil {
		return nil, reverseProxyError("creating listener", err)
	}

	server := DefaultReverseProxyServer(reverseProxy)
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
) (*ReverseProxy, error) {
	dialContext, err := NewDialContext(platform)
	if err != nil {
		return nil, reverseProxyError("creating dialer", err)
	}
	return NewReverseProxyTLSWithDialContext(ctx, dialContext, addr, targetAddr)
}

func NewReverseProxyTLSWithDialContext(
	ctx context.Context,
	dialContext DialContext,
	addr string,
	targetAddr string,
) (*ReverseProxy, error) {
	// TODO: Explicitly use net to create listener to reduce chance of misconfig
	listener, err := NewListener(ctx, NoTEE, Network, addr)
	if err != nil {
		return nil, reverseProxyError("creating listener", err)
	}

	done := make(chan struct{})
	closeFunc := func() error {
		close(done)
		return listener.Close()
	}
	serveFunc := MakeReverseProxyServeFunc(dialContext, listener, targetAddr, done)
	return &ReverseProxy{
		listener:  listener,
		closeFunc: closeFunc,
		serveFunc: serveFunc,
	}, nil
}

func (r *ReverseProxy) Addr() string { return r.listener.Addr().String() }
func (r *ReverseProxy) Close() error { return r.closeFunc() }
func (r *ReverseProxy) Serve() error { return r.serveFunc() }

func MakeReverseProxyServeFunc(
	dialContext DialContext,
	listener net.Listener,
	targetAddr string,
	done chan struct{},
) ServeFunc {
	return func() error {
		for {
			select {
			case <-done:
				return nil
			default:
			}

			clientConn, err := listener.Accept()
			if err != nil {
				return reverseProxyError("accepting connection", err)
			}

			go proxyTLSConn(clientConn, dialContext, targetAddr)
		}
	}
}

// TODO: Can I pass done? Should I pass done?
// TODO: Error channel?
// TODO: Logger?
// TODO: AI code parsed target addr but im pretty sure my custom dial does that
func proxyTLSConn(
	clientConn net.Conn,
	dialContext DialContext,
	targetAddr string,
) {
	defer clientConn.Close()

	// TODO: Pass in a timeout that can be used to create ctx for dial
	serverConn, err := dialContext(context.Background(), Network, targetAddr)
	if err != nil {
		return
	}
	defer serverConn.Close()

	// Bi-directionally forward between client and server
	go io.Copy(serverConn, clientConn)
	io.Copy(clientConn, serverConn)
}

// TODO: Figure out defaults and move to options
func DefaultReverseProxyServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		MaxHeaderBytes:    DefaultMaxHeaderBytes,
		IdleTimeout:       DefaultIdleTimeout,
		ReadHeaderTimeout: DefaultReadHeaderTimeout,
		ReadTimeout:       DefaultReadTimeout,
		WriteTimeout:      DefaultWriteTimeout,
	}
}
