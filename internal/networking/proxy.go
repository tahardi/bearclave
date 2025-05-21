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

	"github.com/tahardi/bearclave/internal/setup"
)

type Proxy struct {
	proxy *httputil.ReverseProxy
}

func NewProxy(
	platform setup.Platform,
	route string,
	cid int,
	port int,
) (*Proxy, error) {
	switch platform {
	case setup.Nitro:
		return NewVSocketProxy(route, cid, port)
	case setup.SEV:
		return NewSocketProxy(route, port)
	case setup.TDX:
		return NewSocketProxy(route, port)
	case setup.NoTEE:
		return NewSocketProxy(route, port)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}

func NewSocketProxy(route string, port int) (*Proxy, error) {
	return NewProxyWithTransport(route, port, &http.Transport{})
}

func NewVSocketProxy(route string, cid int, port int) (*Proxy, error) {
	vsockDialer := func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := vsock.Dial(uint32(cid), uint32(port), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to dial vsock: %v", err)
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
		return nil, fmt.Errorf("Failed to parse target URL: %w", err)
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

type MultiProxy struct {
	proxies map[string]*Proxy
}

func NewMultiProxy(
	platform setup.Platform,
	routes []string,
	cids []int,
	ports []int,
) (*MultiProxy, error) {
	switch platform {
	case setup.Nitro:
		return NewVSocketMultiProxy(routes, cids, ports)
	case setup.SEV:
		return NewSocketMultiProxy(routes, ports)
	case setup.TDX:
		return NewSocketMultiProxy(routes, ports)
	case setup.NoTEE:
		return NewSocketMultiProxy(routes, ports)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}

func NewSocketMultiProxy(routes []string, ports []int) (*MultiProxy, error) {
	switch {
	case len(routes) == 0:
		return nil, fmt.Errorf("no routes provided")
	case len(ports) == 0:
		return nil, fmt.Errorf("no ports provided")
	case len(routes) != len(ports):
		return nil, fmt.Errorf("number of routes and ports must be equal")
	}

	proxies := make(map[string]*Proxy, len(ports))
	for i, route := range routes {
		proxy, err := NewSocketProxy(route, ports[i])
		if err != nil {
			return nil, fmt.Errorf(
				"failed to create proxy for route '%s' and port %d: %w",
				route,
				ports[i],
				err,
			)
		}
		proxies[route] = proxy
	}
	return &MultiProxy{proxies: proxies}, nil
}

func NewVSocketMultiProxy(
	routes []string,
	cids []int,
	ports []int,
) (*MultiProxy, error) {
	switch {
	case len(routes) == 0:
		return nil, fmt.Errorf("no routes provided")
	case len(cids) == 0:
		return nil, fmt.Errorf("no cids provided")
	case len(ports) == 0:
		return nil, fmt.Errorf("no ports provided")
	case len(routes) != len(cids) && len(cids) != len(ports):
		return nil, fmt.Errorf("number of routes, cids, and ports must be equal")
	}

	proxies := make(map[string]*Proxy, len(ports))
	for i, route := range routes {
		proxy, err := NewVSocketProxy(route, cids[i], ports[i])
		if err != nil {
			return nil, fmt.Errorf(
				"failed to create proxy for route '%s' and cid,port (%d,%d): %w",
				route,
				cids[i],
				ports[i],
				err,
			)
		}
		proxies[route] = proxy
	}
	return &MultiProxy{proxies: proxies}, nil
}

func (m *MultiProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for route, proxy := range m.proxies {
		if strings.HasPrefix(r.URL.Path, route) {
			proxy.ServeHTTP(w, r)
			return
		}
	}

	http.Error(w, "No matching route found", http.StatusNotFound)
}
