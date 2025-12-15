package bearclave

import (
	"fmt"

	"github.com/tahardi/bearclave/internal/clock"
)

var ErrTimer = clock.ErrTimer

type Timer = clock.Timer

func NewTimer(platform Platform) (Timer, error) {
	switch platform {
	case Nitro:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, platform)
	case SEV:
		return clock.NewTSCTimer()
	case TDX:
		return clock.NewTSCTimer()
	case NoTEE:
		return clock.NewTSCTimer()
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, platform)
	}
}
