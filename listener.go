package bearclave

import (
	"context"
	"fmt"
	"net"

	"github.com/tahardi/bearclave/internal/networking"
)

var ErrListener = networking.ErrListener

func NewListener(
	ctx context.Context,
	platform Platform,
	network string,
	addr string,
	options ...ListenerOption,
) (net.Listener, error) {
	switch platform {
	case Nitro:
		return networking.NewVSocketListener(ctx, network, addr, options...)
	case SEV:
		return networking.NewSocketListener(ctx, network, addr, options...)
	case TDX:
		return networking.NewSocketListener(ctx, network, addr, options...)
	case NoTEE:
		return networking.NewSocketListener(ctx, network, addr, options...)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, platform)
	}
}

type ListenerOption = networking.ListenerOption
type ListenerOptions = networking.ListenerOptions

var (
	WithListenControl         = networking.WithListenControl
	WithListenKeepAlive       = networking.WithListenKeepAlive
	WithListenKeepAliveConfig = networking.WithListenKeepAliveConfig
)
