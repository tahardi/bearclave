package attestation

import (
	"fmt"

	"github.com/tahardi/bearclave/internal/attestation/nitro"
	"github.com/tahardi/bearclave/internal/attestation/notee"
	"github.com/tahardi/bearclave/internal/attestation/sev"
	"github.com/tahardi/bearclave/internal/attestation/tdx"
	"github.com/tahardi/bearclave/internal/setup"
)

type Attester interface {
	// TODO: Update to take [64]byte?
	Attest(userdata []byte) (attestation []byte, err error)
}

func NewAttester(platform setup.Platform) (Attester, error) {
	switch platform {
	case setup.Nitro:
		return nitro.NewAttester()
	case setup.SEV:
		return sev.NewAttester()
	case setup.TDX:
		return tdx.NewAttester()
	case setup.NoTEE:
		return notee.NewAttester()
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}
