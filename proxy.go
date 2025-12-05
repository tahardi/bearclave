package bearclave

import (
	"fmt"

	"github.com/tahardi/bearclave/internal/networking"
)

type Proxy = networking.Proxy

func NewProxy(
	platform Platform,
	route string,
	cid int,
	port int,
) (*Proxy, error) {
	switch platform {
	case Nitro:
		return networking.NewVSocketProxy(route, cid, port)
	case SEV:
		return networking.NewSocketProxy(route, port)
	case TDX:
		return networking.NewSocketProxy(route, port)
	case NoTEE:
		return networking.NewSocketProxy(route, port)
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}
