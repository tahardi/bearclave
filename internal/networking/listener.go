package networking

import (
	"fmt"
	"net"

	"github.com/mdlayher/vsock"
)

func NewSocketListener(network string, addr string) (net.Listener, error) {
	listener, err := net.Listen(network, addr)
	if err != nil {
		return nil, fmt.Errorf(
			"creating %s socket listener on %s: %w",
			network,
			addr,
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
	var cid, port int
	n, err := fmt.Sscanf(addr, "%d:%d", &cid, &port)
	switch {
	case err != nil:
		return 0, 0, fmt.Errorf("invalid format '%s': %w", addr, err)
	case n != 2:
		return 0, 0, fmt.Errorf("expected '2' got '%d", n)
	}
	return uint32(cid), uint32(port), nil
}