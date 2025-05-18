package attestation

import (
	"fmt"

	"github.com/tahardi/bearclave/internal/attestation/nitro"
	"github.com/tahardi/bearclave/internal/attestation/notee"
	"github.com/tahardi/bearclave/internal/attestation/sev"
	"github.com/tahardi/bearclave/internal/attestation/tdx"
	"github.com/tahardi/bearclave/internal/setup"
)

type Verifier interface {
	Verify(attestation []byte) (userdata []byte, err error)
}

func NewVerifier(platform setup.Platform) (Verifier, error) {
	switch platform {
	case setup.Nitro:
		return nitro.NewVerifier()
	case setup.SEV:
		return sev.NewVerifier()
	case setup.TDX:
		return tdx.NewVerifier()
	case setup.NoTEE:
		return notee.NewVerifier()
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}
