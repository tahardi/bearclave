package networking

import (
	"net"

	"github.com/mdlayher/vsock"
)

type Dialer func(network string, addr string) (net.Conn, error)

func NewSocketDialer() (Dialer, error) {
	return func(network string, addr string) (net.Conn, error) {
		parsedAddr, err := ParseSocketAddr(addr)
		if err != nil {
			return nil, dialerError("", err)
		}
		return net.Dial(network, parsedAddr)
	}, nil
}

func NewVSocketDialer() (Dialer, error) {
	return func(_ string, addr string) (net.Conn, error) {
		cid, port, err := ParseVSocketAddr(addr)
		if err != nil {
			return nil, dialerError("", err)
		}
		return vsock.Dial(cid, port, nil)
	}, nil
}
