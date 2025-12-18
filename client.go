package bearclave

import (
	"net/http"

	"github.com/tahardi/bearclave/internal/networking"
)

func NewProxiedClient(
	platform Platform,
	proxyAddr string,
) (*http.Client, error) {
	switch platform {
	case Nitro:
		return networking.NewProxiedVSocketClient(proxyAddr)
	case SEV:
		return networking.NewProxiedSocketClient(proxyAddr)
	case TDX:
		return networking.NewProxiedSocketClient(proxyAddr)
	case NoTEE:
		return networking.NewProxiedSocketClient(proxyAddr)
	default:
		return nil, ErrUnsupportedPlatform
	}
}
