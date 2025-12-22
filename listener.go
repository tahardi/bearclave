package bearclave

import (
	"github.com/tahardi/bearclave/internal/networking"
)

var (
	NewSocketListener  = networking.NewSocketListener
	NewVSocketListener = networking.NewVSocketListener
)

type ListenerOption = networking.ListenerOption
type ListenerOptions = networking.ListenerOptions

var (
	WithListenControl         = networking.WithListenControl
	WithListenKeepAlive       = networking.WithListenKeepAlive
	WithListenKeepAliveConfig = networking.WithListenKeepAliveConfig
)
