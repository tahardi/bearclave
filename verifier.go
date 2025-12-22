package bearclave

import (
	"github.com/tahardi/bearclave/internal/attestation"
)

type Verifier = attestation.Verifier

var (
	NewNitroVerifier = attestation.NewNitroVerifier
	NewSEVVerifier   = attestation.NewSEVVerifier
	NewTDXVerifier   = attestation.NewTDXVerifier
	NewNoTEEVerifier = attestation.NewNoTEEVerifier
)

type VerifyResult = attestation.VerifyResult
type VerifyOption = attestation.VerifyOption
type VerifyOptions = attestation.VerifyOptions

var (
	WithVerifyDebug       = attestation.WithVerifyDebug
	WithVerifyMeasurement = attestation.WithVerifyMeasurement
	WithVerifyTimestamp   = attestation.WithVerifyTimestamp
	WithVerifyNonce       = attestation.WithVerifyVerifyNonce
)
