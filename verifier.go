package bearclave

import (
	"fmt"

	"github.com/tahardi/bearclave/internal/attestation"
)

var (
	ErrVerifier            = attestation.ErrVerifier
	ErrVerifierDebugMode   = attestation.ErrVerifierDebugMode
	ErrVerifierMeasurement = attestation.ErrVerifierMeasurement
	ErrVerifierNonce       = attestation.ErrVerifierNonce
	ErrVerifierTimestamp   = attestation.ErrVerifierTimestamp
	WithDebug              = attestation.WithDebug
	WithMeasurement        = attestation.WithMeasurement
	WithTimestamp          = attestation.WithTimestamp
	WithVerifyNonce        = attestation.WithVerifyNonce
)

type Verifier = attestation.Verifier
type VerifyResult = attestation.VerifyResult
type VerifyOption = attestation.VerifyOption
type VerifyOptions = attestation.VerifyOptions

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
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, platform)
	}
}
