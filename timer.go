package bearclave

import (
	"github.com/tahardi/bearclave/internal/clock"
)

var ErrTimer = clock.ErrTimer

type Timer = clock.Timer

var NewTSCTimer = clock.NewTSCTimer
