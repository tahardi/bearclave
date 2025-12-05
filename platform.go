package bearclave

import "github.com/tahardi/bearclave/internal/setup"

type Platform = setup.Platform

const (
	Nitro Platform = setup.Nitro
	SEV   Platform = setup.SEV
	TDX   Platform = setup.TDX
	NoTEE Platform = setup.NoTEE
)

