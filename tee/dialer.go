package tee

import (
	"fmt"

	"github.com/tahardi/bearclave"
)

var ErrDialContext = bearclave.ErrDialContext

type DialContext = bearclave.DialContext

func NewDialContext(
	platform Platform,
	options ...DialerOption,
) (DialContext, error) {
	switch platform {
	case Nitro:
		return bearclave.NewVSocketDialContext(options...)
	case SEV:
		return bearclave.NewSocketDialContext(options...)
	case TDX:
		return bearclave.NewSocketDialContext(options...)
	case NoTEE:
		return bearclave.NewSocketDialContext(options...)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, platform)
	}
}

type DialerOption = bearclave.DialerOption
type DialerOptions = bearclave.DialerOptions

var (
	WithDialControl   = bearclave.WithDialControl
	WithDialKeepAlive = bearclave.WithDialKeepAlive
	WithDialLocalAddr = bearclave.WithDialLocalAddr
	WithDialTimeout   = bearclave.WithDialTimeout
)
