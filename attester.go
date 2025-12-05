package bearclave

import (
	"fmt"

	"github.com/tahardi/bearclave/internal/attestation"
)

type Attester = attestation.Attester

func NewAttester(platform Platform) (Attester, error) {
	switch platform {
	case Nitro:
		return attestation.NewNitroAttester()
	case SEV:
		return attestation.NewSEVAttester()
	case TDX:
		return attestation.NewTDXAttester()
	case NoTEE:
		return attestation.NewNoTEEAttester()
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}
