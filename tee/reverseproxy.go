package tee

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/tahardi/bearclave"
)

func NewReverseProxy(
	platform bearclave.Platform,
	targetAddr string,
	route string,
) (*httputil.ReverseProxy, error) {
	dialer, err := bearclave.NewDialer(platform)
	if err != nil {
		return nil, fmt.Errorf("creating dialer: %w", err)
	}
	return NewReverseProxyWithDialer(dialer, targetAddr, route)
}

func NewReverseProxyWithDialer(
	dialer bearclave.Dialer,
	targetAddr string,
	route string,
) (*httputil.ReverseProxy, error) {
	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		dataChan := make(chan net.Conn, 1)
		errChan := make(chan error, 1)
		go func() {
			conn, err := dialer(network, addr)
			if err != nil {
				errChan <- fmt.Errorf("dialing '%s': %w", addr, err)
			}
			dataChan <- conn
		}()
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("deadline exceeded or context cancelled: %w", ctx.Err())
		case err := <-errChan:
			return nil, err
		case data := <-dataChan:
			return data, nil
		}
	}
	transport := &http.Transport{DialContext: dialContext}
	return NewReverseProxyWithTransport(transport, targetAddr, route)
}

func NewReverseProxyWithTransport(
	transport *http.Transport,
	targetAddr string,
	route string,
) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(targetAddr)
	if err != nil {
		return nil, fmt.Errorf("parsing target URL: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = transport

	// Customize the Director to strip the path prefix
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, route)
	}
	return proxy, nil
}
