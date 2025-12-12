package bearclave

import (
	"fmt"

	"github.com/tahardi/bearclave/internal/networking"
)

var (
	ErrDialContext     = networking.ErrDialContext
	WithControlContext = networking.WithControlContext
	WithKeepAlive      = networking.WithKeepAlive
	WithLocalAddr      = networking.WithLocalAddr
	WithTimeout        = networking.WithTimeout
)

type DialContext = networking.DialContext
type DialerOption = networking.DialerOption
type DialerOptions = networking.DialerOptions

func NewDialContext(
	platform Platform,
	options ...DialerOption,
) (networking.DialContext, error) {
	switch platform {
	case Nitro:
		return networking.NewVSocketDialContext(options...)
	case SEV:
		return networking.NewSocketDialContext(options...)
	case TDX:
		return networking.NewSocketDialContext(options...)
	case NoTEE:
		return networking.NewSocketDialContext(options...)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, platform)
	}
}
