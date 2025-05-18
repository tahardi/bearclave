package networking

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/mdlayher/vsock"

	"github.com/tahardi/bearclave/internal/setup"
)

type Proxy struct {
	proxy *httputil.ReverseProxy
}

func NewProxy(
	platform setup.Platform,
	cid int,
	port int,
) (*Proxy, error) {
	switch platform {
	case setup.Nitro:
		return NewVSocketProxy(cid, port)
	case setup.SEV:
		return NewSocketProxy(port)
	case setup.TDX:
		return NewSocketProxy(port)
	case setup.NoTEE:
		return NewSocketProxy(port)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}

func NewSocketProxy(port int) (*Proxy, error) {
	return NewProxyWithTransport(port, &http.Transport{})
}

func NewVSocketProxy(cid int, port int) (*Proxy, error) {
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
	return NewProxyWithTransport(port, transport)
}

func NewProxyWithTransport(
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
	return &Proxy{
		proxy: proxy,
	}, nil
}

func (p *Proxy) Handler() http.Handler {
	return p.proxy
}
