package tee

import (
	"github.com/tahardi/bearclave"
)

var ErrTimer = bearclave.ErrTimer

type Timer = bearclave.Timer

func NewTimer(platform Platform) (Timer, error) {
	switch platform {
	case Nitro:
		return bearclave.NewTSCTimer()
	case SEV:
		return bearclave.NewTSCTimer()
	case TDX:
		return bearclave.NewTSCTimer()
	case NoTEE:
		return bearclave.NewTSCTimer()
	default:
		return nil, teeErrorUnsupportedPlatform(string(platform), nil)
	}
}
