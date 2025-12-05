package tee

import (
	"fmt"

	"github.com/tahardi/bearclave"
	"github.com/tahardi/bearclave/internal/networking"
)

type Proxy = networking.Proxy

func NewProxy(
	platform bearclave.Platform,
	route string,
	cid int,
	port int,
) (*Proxy, error) {
	switch platform {
	case bearclave.Nitro:
		return networking.NewVSocketProxy(route, cid, port)
	case bearclave.SEV:
		return networking.NewSocketProxy(route, port)
	case bearclave.TDX:
		return networking.NewSocketProxy(route, port)
	case bearclave.NoTEE:
		return networking.NewSocketProxy(route, port)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}
