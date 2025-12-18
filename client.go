package bearclave

import (
	"github.com/tahardi/bearclave/internal/networking"
)

var (
	NewProxiedSocketClient = networking.NewProxiedSocketClient
	NewProxiedVSocketClient = networking.NewProxiedVSocketClient
)
