package networking

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/mdlayher/vsock"
)

const NumVSockAddrFields = 2

func sanitizeAddr(addr string) (string, error) {
	if !strings.Contains(addr, "://") {
		return addr, nil
	}

	u, err := url.Parse(addr)
	if err != nil {
		return "", fmt.Errorf("parsing URL: %w", err)
	}
	return u.Host, nil
}

func NewSocketListener(network string, addr string) (net.Listener, error) {
	sanitizedAddr, err := sanitizeAddr(addr)
	if err != nil {
		return nil, fmt.Errorf("sanitizing addr: %w", err)
	}

	listener, err := net.Listen(network, sanitizedAddr)
	if err != nil {
		return nil, fmt.Errorf(
			"creating %s socket listener on %s: %w",
			network,
			sanitizedAddr,
			err,
		)
	}
	return listener, nil
}

func NewVSocketListener(_ string, addr string) (net.Listener, error) {
	cid, port, err := ParseVSocketAddr(addr)
	if err != nil {
		return nil, fmt.Errorf("parsing vsocket addr: %w", err)
	}

	listener, err := vsock.ListenContextID(cid, port, nil)
	if err != nil {
		return nil, fmt.Errorf(
			"creating vsocket listener on '%s': %w",
			addr,
			err,
		)
	}
	return listener, nil
}

func ParseVSocketAddr(addr string) (uint32, uint32, error) {
	// Handle URLs like "http://cid:port" by stripping the scheme
	if idx := strings.Index(addr, "://"); idx != -1 {
		addr = addr[idx+3:]
	}

	var cid, port int
	n, err := fmt.Sscanf(addr, "%d:%d", &cid, &port)
	switch {
	case err != nil:
		return 0, 0, fmt.Errorf("invalid format '%s': %w", addr, err)
	case n != NumVSockAddrFields:
		return 0, 0, fmt.Errorf("expected '2' got '%d", n)
	}
	return uint32(cid), uint32(port), nil
}