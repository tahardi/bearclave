package tee

import (
	"fmt"
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
	dialer, err := bearclave.NewDialContext(platform)
	if err != nil {
		return nil, fmt.Errorf("creating dialer: %w", err)
	}
	return NewReverseProxyWithDialContext(dialer, targetAddr, route)
}

func NewReverseProxyWithDialContext(
	dialContext bearclave.DialContext,
	targetAddr string,
	route string,
) (*httputil.ReverseProxy, error) {
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
