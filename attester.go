package bearclave

import (
	"github.com/tahardi/bearclave/internal/attestation"
)

var (
	ErrAttester                = attestation.ErrAttester
	ErrAttesterUserDataTooLong = attestation.ErrAttesterUserDataTooLong
)

type Attester = attestation.Attester

var (
	NewNitroAttester = attestation.NewNitroAttester
	NewSEVAttester   = attestation.NewSEVAttester
	NewTDXAttester   = attestation.NewTDXAttester
	NewNoTEEAttester = attestation.NewNoTEEAttester
)

type AttestResult = attestation.AttestResult
type AttestOption = attestation.AttestOption
type AttestOptions = attestation.AttestOptions

var (
	WithAttestNonce     = attestation.WithAttestNonce
	WithAttestUserData  = attestation.WithAttestUserData
)
