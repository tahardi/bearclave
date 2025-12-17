package bearclave

import (
	"fmt"

	"github.com/tahardi/bearclave/internal/attestation"
)

var (
	ErrAttester                = attestation.ErrAttester
	ErrAttesterUserDataTooLong = attestation.ErrAttesterUserDataTooLong
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
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, platform)
	}
}

type AttestResult = attestation.AttestResult
type AttestOption = attestation.AttestOption
type AttestOptions = attestation.AttestOptions

var (
	WithAttestNonce     = attestation.WithAttestNonce
	WithAttestPublicKey = attestation.WithAttestPublicKey
	WithAttestUserData  = attestation.WithAttestUserData
)
