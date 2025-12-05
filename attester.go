package bearclave

import (
	"github.com/tahardi/bearclave/internal/attestation"
)

type Attester = attestation.Attester

func NewAttester(platform Platform) (Attester, error) {
	return attestation.NewAttester(platform)
}
