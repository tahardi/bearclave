package bearclave

import (
	"context"
	"fmt"
	"net"

	"github.com/tahardi/bearclave/internal/networking"
)

var (
	ErrListener               = networking.ErrListener
	WithListenControl         = networking.WithListenControl
	WithListenKeepAlive       = networking.WithListenKeepAlive
	WithListenKeepAliveConfig = networking.WithListenKeepAliveConfig
)

func NewListener(
	ctx context.Context,
	platform Platform,
	network string,
	addr string,
) (net.Listener, error) {
	switch platform {
	case Nitro:
		return networking.NewVSocketListener(ctx, network, addr)
	case SEV:
		return networking.NewSocketListener(ctx, network, addr)
	case TDX:
		return networking.NewSocketListener(ctx, network, addr)
	case NoTEE:
		return networking.NewSocketListener(ctx, network, addr)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, platform)
	}
}
