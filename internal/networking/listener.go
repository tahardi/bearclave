package networking

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/mdlayher/vsock"
)

const NumVSockAddrFields = 2

func NewSocketListener(network string, addr string) (net.Listener, error) {
	parsedAddr, err := ParseSocketAddr(addr)
	if err != nil {
		return nil, listenerError("", err)
	}

	listener, err := net.Listen(network, parsedAddr)
	if err != nil {
		msg := fmt.Sprintf(
			"creating %s socket listener on %s",
			network,
			parsedAddr,
		)
		return nil, listenerError(msg, err)
	}
	return listener, nil
}

func NewVSocketListener(_ string, addr string) (net.Listener, error) {
	cid, port, err := ParseVSocketAddr(addr)
	if err != nil {
		return nil, listenerError("", err)
	}

	listener, err := vsock.ListenContextID(cid, port, nil)
	if err != nil {
		msg := fmt.Sprintf("creating vsocket listener on '%s'", addr)
		return nil, listenerError(msg, err)
	}
	return listener, nil
}

func ParseSocketAddr(addr string) (string, error) {
	if !strings.Contains(addr, "://") {
		return addr, nil
	}

	u, err := url.Parse(addr)
	if err != nil {
		return "", fmt.Errorf("%w: parsing URL: %w", ErrSocketParseAddr, err)
	}
	return u.Host, nil
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
		msg := fmt.Sprintf("expected format 'cid:port' got '%s'", addr)
		return 0, 0, fmt.Errorf("%w: %s: %w", ErrVSocketParseAddr, msg, err)
	case n != NumVSockAddrFields:
		msg := fmt.Sprintf("expected 2 fields got %d", n)
		return 0, 0, fmt.Errorf("%w: %s", ErrVSocketParseAddr, msg)
	}
	return uint32(cid), uint32(port), nil
}
