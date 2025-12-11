package bearclave

import (
	"fmt"
	"net"

	"github.com/tahardi/bearclave/internal/networking"
)

var ErrListener = networking.ErrListener

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
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, platform)
	}
}
