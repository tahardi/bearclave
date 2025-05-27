package networking

import (
	"github.com/tahardi/bearclave/internal/networking"
	"github.com/tahardi/bearclave/pkg/setup"
)

type Proxy = networking.Proxy

func NewProxy(
	platform setup.Platform,
	route string,
	cid int,
	port int,
) (*Proxy, error) {
	return networking.NewProxy(platform, route, cid, port)
}

type MultiProxy = networking.MultiProxy

func NewMultiProxy(
	platform setup.Platform,
	routes []string,
	cids []int,
	ports []int,
) (*MultiProxy, error) {
	return networking.NewMultiProxy(platform, routes, cids, ports)
}
