package attestation

import (
	"github.com/tahardi/bearclave/internal/attestation"
	"github.com/tahardi/bearclave/pkg/setup"
)

type Verifier = attestation.Verifier

func NewVerifier(platform setup.Platform) (Verifier, error) {
	return attestation.NewVerifier(platform)
}
