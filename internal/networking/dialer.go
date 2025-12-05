package networking

import (
	"fmt"
	"net"

	"github.com/mdlayher/vsock"
)

type Dialer func(network string, addr string) (net.Conn, error)

func NewSocketDialer() (Dialer, error) {
	return net.Dial, nil
}

func NewVSocketDialer() (Dialer, error) {
	return func(network string, addr string) (net.Conn, error) {
		cid, port, err := ParseVSocketAddr(addr)
		if err != nil {
			return nil, fmt.Errorf("parsing vsocket addr: %w", err)
		}
		return vsock.Dial(cid, port, nil)
	}, nil
}
