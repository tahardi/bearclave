package bearclave

import (
	"fmt"

	"github.com/tahardi/bearclave/internal/networking"
)

var ErrDialer = networking.ErrDialer

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
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, platform)
	}
}
