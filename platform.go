package bearclave

type Platform string

const (
	Nitro Platform = "nitro"
	SEV   Platform = "sev"
	TDX   Platform = "tdx"
	NoTEE Platform = "notee"
)

