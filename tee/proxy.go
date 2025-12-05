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
	targetPort int,
	route string,
) (*httputil.ReverseProxy, error) {
	dialer, err := bearclave.NewDialer(platform)
	if err != nil {
		return nil, fmt.Errorf("creating dialer: %w", err)
	}
	return NewReverseProxyWithDialer(dialer, targetPort, route)
}

func NewReverseProxyWithDialer(
	dialer bearclave.Dialer,
	targetPort int,
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
	return NewReverseProxyWithTransport(transport, targetPort, route)
}

func NewReverseProxyWithTransport(
	transport *http.Transport,
	targetPort int,
	route string,
) (*httputil.ReverseProxy, error) {
	addr := fmt.Sprintf("http://127.0.0.1:%d", targetPort)
	targetURL, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("parsing target URL: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = transport

	// Customize the Director to handle path rewriting
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Rewrite the path by stripping the prefix
		req.URL.Path = strings.TrimPrefix(req.URL.Path, route)
	}
	return proxy, nil
}
