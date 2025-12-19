package tee

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func NewReverseProxy(
	ctx context.Context,
	platform Platform,
	network string,
	proxyAddr string,
	targetAddr string,
	route string,
	options ...ServerOption,
) (*Server, error) {
	dialer, err := NewDialContext(platform)
	if err != nil {
		return nil, fmt.Errorf("creating dialer: %w", err)
	}
	return NewReverseProxyWithDialContext(
		ctx,
		dialer,
		network,
		proxyAddr,
		targetAddr,
		route,
		options...,
	)
}

func NewReverseProxyWithDialContext(
	ctx context.Context,
	dialContext DialContext,
	network string,
	proxyAddr string,
	targetAddr string,
	route string,
	options ...ServerOption,
) (*Server, error) {
	targetURL, err := url.Parse(targetAddr)
	if err != nil {
		return nil, fmt.Errorf("parsing target URL: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = &http.Transport{DialContext: dialContext}

	// Customize the Director to strip the path prefix
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, route)
	}

	// NOTE: We assume that reverse proxies are only ever run (1) outside a
	// Nitro Enclave or (2) within an SEV-SNP/TDX enclave. For Nitro, the
	// DialContext must use vsock to send to the Enclave, but it must use a
	// normal socket to listen for inbound connections from remote clients.
	// This is why we enforce a normal socket listener here. Even though a
	// reverse proxy runs within the trusted zone for SEV-SNP and TDX, they
	// both use normal sockets, so using a socket listener here is fine.
	listener, err := NewListener(ctx, NoTEE, network, proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("creating listener: %w", err)
	}
	return NewServerWithListener(listener, proxy, options...)
}
