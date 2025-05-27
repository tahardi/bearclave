package setup

import "github.com/tahardi/bearclave/internal/setup"

type Platform = setup.Platform

const (
	Nitro = setup.Nitro
	SEV   = setup.SEV
	TDX   = setup.TDX
	NoTEE = setup.NoTEE
)
