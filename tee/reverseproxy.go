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
	// Target addresses for nitro are of the form "cid:port", which does not
	// work with url.Parse. Extract the port and build a localhost URL since
	// we assume the proxy and enclave run on the same machine.
	tokens := strings.Split(targetAddr, ":")
	if len(tokens) < 2 {
		return nil, fmt.Errorf("invalid target address: %s", targetAddr)
	}

	targetURL, err := url.Parse(fmt.Sprintf("127.0.0.1:%s", tokens[len(tokens)-1]))
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
