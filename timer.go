package bearclave

import (
	"github.com/tahardi/bearclave/internal/clock"
)

type Timer = clock.Timer

var NewTSCTimer = clock.NewTSCTimer
