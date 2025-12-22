package bearclave

import (
	"github.com/tahardi/bearclave/internal/attestation"
	"github.com/tahardi/bearclave/internal/clock"
	"github.com/tahardi/bearclave/internal/networking"
)

var (
	ErrAttester            = attestation.ErrAttester
	ErrAttesterUserData    = attestation.ErrAttesterUserData
	ErrDialContext         = networking.ErrDialContext
	ErrListener            = networking.ErrListener
	ErrTimer               = clock.ErrTimer
	ErrVerifier            = attestation.ErrVerifier
	ErrVerifierDebugMode   = attestation.ErrVerifierDebugMode
	ErrVerifierMeasurement = attestation.ErrVerifierMeasurement
	ErrVerifierNonce       = attestation.ErrVerifierNonce
	ErrVerifierTimestamp   = attestation.ErrVerifierTimestamp
)
