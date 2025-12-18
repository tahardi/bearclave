package bearclave

import (
	"github.com/tahardi/bearclave/internal/networking"
)

var ErrDialContext = networking.ErrDialContext

type DialContext = networking.DialContext

var (
	NewSocketDialContext  = networking.NewSocketDialContext
	NewVSocketDialContext = networking.NewVSocketDialContext
)

type DialerOption = networking.DialerOption
type DialerOptions = networking.DialerOptions

var (
	WithDialControl   = networking.WithDialControl
	WithDialKeepAlive = networking.WithDialKeepAlive
	WithDialLocalAddr = networking.WithDialLocalAddr
	WithDialTimeout   = networking.WithDialTimeout
)
