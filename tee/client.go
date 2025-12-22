package tee

import (
	"net/http"

	"github.com/tahardi/bearclave"
)

func NewProxiedClient(
	platform Platform,
	proxyAddr string,
) (*http.Client, error) {
	switch platform {
	case Nitro:
		return bearclave.NewProxiedVSocketClient(proxyAddr)
	case SEV:
		return bearclave.NewProxiedSocketClient(proxyAddr)
	case TDX:
		return bearclave.NewProxiedSocketClient(proxyAddr)
	case NoTEE:
		return bearclave.NewProxiedSocketClient(proxyAddr)
	default:
		return nil, unsupportedPlatformError(string(platform), nil)
	}
}
