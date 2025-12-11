package networking

import (
	"fmt"
	"net"

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
