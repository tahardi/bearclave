package bearclave

import (
	"fmt"
	"net"

	"github.com/tahardi/bearclave/internal/networking"
)

type Dialer = networking.Dialer

func NewDialer(platform Platform) (networking.Dialer, error) {
	switch platform {
	case Nitro:
		return networking.NewVSocketDialer()
	case SEV:
		return networking.NewSocketDialer()
	case TDX:
		return networking.NewSocketDialer()
	case NoTEE:
		return networking.NewSocketDialer()
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}

func NewListener(platform Platform, network string, addr string) (net.Listener, error) {
	switch platform {
	case Nitro:
		return networking.NewVSocketListener(network, addr)
	case SEV:
		return networking.NewSocketListener(network, addr)
	case TDX:
		return networking.NewSocketListener(network, addr)
	case NoTEE:
		return networking.NewSocketListener(network, addr)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}
