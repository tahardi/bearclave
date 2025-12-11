package bearclave

import "errors"

type Platform string

const (
	Nitro Platform = "nitro"
	SEV   Platform = "sev"
	TDX   Platform = "tdx"
	NoTEE Platform = "notee"
)

var ErrUnsupportedPlatform = errors.New("unsupported platform")