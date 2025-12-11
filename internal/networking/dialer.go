package networking

import (
	"errors"
	"fmt"
	"net"

	"github.com/mdlayher/vsock"
)

var ErrDialer = errors.New("dialer")

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

func wrapDialerError(dialerErr error, msg string, err error) error {
	switch {
	case msg == "" && err == nil:
		return dialerErr
	case msg != "" && err != nil:
		return fmt.Errorf("%w: %s: %w", dialerErr, msg, err)
	case msg != "":
		return fmt.Errorf("%w: %s", dialerErr, msg)
	default:
		return fmt.Errorf("%w: %w", dialerErr, err)
	}
}

func dialerError(msg string, err error) error {
	return wrapDialerError(ErrDialer, msg, err)
}
