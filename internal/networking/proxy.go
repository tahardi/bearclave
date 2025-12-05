package networking

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/mdlayher/vsock"
)

type Proxy struct {
	proxy *httputil.ReverseProxy
}

func NewSocketProxy(route string, port int) (*Proxy, error) {
	return NewProxyWithTransport(route, port, &http.Transport{})
}

func NewVSocketProxy(route string, cid int, port int) (*Proxy, error) {
	vsockDialer := func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := vsock.Dial(uint32(cid), uint32(port), nil)
		if err != nil {
			return nil, fmt.Errorf("dialing vsock: %v", err)
		}
		return conn, nil
	}

	transport := &http.Transport{
		DialContext: vsockDialer,
	}
	return NewProxyWithTransport(route, port, transport)
}

func NewProxyWithTransport(
	route string,
	port int,
	transport *http.Transport,
) (*Proxy, error) {
	addr := fmt.Sprintf("http://127.0.0.1:%d", port)
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

	return &Proxy{
		proxy: proxy,
	}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}
