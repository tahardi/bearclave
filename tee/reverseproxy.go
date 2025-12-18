package tee

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/tahardi/bearclave"
)

func NewReverseProxy(
	ctx context.Context,
	platform bearclave.Platform,
	network string,
	proxyAddr string,
	targetAddr string,
	route string,
) (*Server, error) {
	dialer, err := bearclave.NewDialContext(platform)
	if err != nil {
		return nil, fmt.Errorf("creating dialer: %w", err)
	}
	return NewReverseProxyWithDialContext(
		ctx,
		platform,
		dialer,
		network,
		proxyAddr,
		targetAddr,
		route,
	)
}

func NewReverseProxyWithDialContext(
	ctx context.Context,
	platform bearclave.Platform,
	dialContext bearclave.DialContext,
	network string,
	proxyAddr string,
	targetAddr string,
	route string,
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
	return NewServer(ctx, platform, network, proxyAddr, proxy)
}
