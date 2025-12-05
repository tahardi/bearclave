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

type AttestResult = attestation.AttestResult
type AttestOption = attestation.AttestOption
type AttestOptions = attestation.AttestOptions

var (
	WithAttestNonce = attestation.WithAttestNonce
	WithPublicKey   = attestation.WithPublicKey
	WithUserData    = attestation.WithUserData
)

type Verifier = attestation.Verifier

func NewVerifier(platform Platform) (Verifier, error) {
	switch platform {
	case Nitro:
		return attestation.NewNitroVerifier()
	case SEV:
		return attestation.NewSEVVerifier()
	case TDX:
		return attestation.NewTDXVerifier()
	case NoTEE:
		return attestation.NewNoTEEVerifier()
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}

type VerifyResult = attestation.VerifyResult
type VerifyOption = attestation.VerifyOption
type VerifyOptions = attestation.VerifyOptions

var (
	WithDebug = attestation.WithDebug
	WithMeasurement = attestation.WithMeasurement
	WithTimestamp = attestation.WithTimestamp
	WithVerifyNonce = attestation.WithVerifyNonce
)