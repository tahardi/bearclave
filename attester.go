package attestation

import (
	"github.com/tahardi/bearclave/internal/attestation"
	"github.com/tahardi/bearclave/pkg/setup"
)

type Attester = attestation.Attester

func NewAttester(platform setup.Platform) (Attester, error) {
	return attestation.NewAttester(platform)
}
