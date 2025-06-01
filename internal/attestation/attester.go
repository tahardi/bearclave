package attestation

import (
	"fmt"

	"github.com/tahardi/bearclave/internal/setup"
)

type Attester interface {
	// TODO: Update to take [64]byte?
	Attest(userdata []byte) (report []byte, err error)
}

func NewAttester(platform setup.Platform) (Attester, error) {
	switch platform {
	case setup.Nitro:
		return NewNitroAttester()
	case setup.SEV:
		return NewSEVAttester()
	case setup.TDX:
		return NewTDXAttester()
	case setup.NoTEE:
		return NewNoTEEAttester()
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}
