package tee

import (
	"context"
	"net"

	"github.com/tahardi/bearclave"
)

func NewListener(
	ctx context.Context,
	platform Platform,
	network string,
	addr string,
	options ...ListenerOption,
) (net.Listener, error) {
	switch platform {
	case Nitro:
		return bearclave.NewVSocketListener(ctx, network, addr, options...)
	case SEV:
		return bearclave.NewSocketListener(ctx, network, addr, options...)
	case TDX:
		return bearclave.NewSocketListener(ctx, network, addr, options...)
	case NoTEE:
		return bearclave.NewSocketListener(ctx, network, addr, options...)
	default:
		return nil, unsupportedPlatformError(string(platform), nil)
	}
}

type ListenerOption = bearclave.ListenerOption
type ListenerOptions = bearclave.ListenerOptions

var (
	WithListenControl         = bearclave.WithListenControl
	WithListenKeepAlive       = bearclave.WithListenKeepAlive
	WithListenKeepAliveConfig = bearclave.WithListenKeepAliveConfig
)
