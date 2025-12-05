package bearclave

import (
	"github.com/tahardi/bearclave/internal/networking"
)

type Proxy = networking.Proxy

func NewProxy(
	platform Platform,
	route string,
	cid int,
	port int,
) (*Proxy, error) {
	return networking.NewProxy(platform, route, cid, port)
}

type MultiProxy = networking.MultiProxy

func NewMultiProxy(
	platform Platform,
	routes []string,
	cids []int,
	ports []int,
) (*MultiProxy, error) {
	return networking.NewMultiProxy(platform, routes, cids, ports)
}
