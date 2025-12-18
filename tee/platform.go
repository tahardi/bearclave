package tee

import "errors"

var ErrUnsupportedPlatform = errors.New("unsupported platform")

type Platform string

const (
	Nitro Platform = "nitro"
	SEV   Platform = "sev"
	TDX   Platform = "tdx"
	NoTEE Platform = "notee"
)
