package networking

import (
	"fmt"
	"net/url"
	"strings"
)

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

	var cid, port uint32
	n, err := fmt.Sscanf(addr, "%d:%d", &cid, &port)
	switch {
	case err != nil:
		msg := fmt.Sprintf("expected format 'cid:port' got '%s'", addr)
		return 0, 0, fmt.Errorf("%w: %s: %w", ErrVSocketParseAddr, msg, err)
	case n != NumVSockAddrFields:
		msg := fmt.Sprintf("expected 2 fields got %d", n)
		return 0, 0, fmt.Errorf("%w: %s", ErrVSocketParseAddr, msg)
	}
	return cid, port, nil
}
