package attestation

import (
	"time"

	"github.com/tahardi/bearclave/internal/attestation"
	"github.com/tahardi/bearclave/pkg/setup"
)

type Verifier = attestation.Verifier

func NewVerifier(platform setup.Platform) (Verifier, error) {
	return attestation.NewVerifier(platform)
}

type VerifyOption = attestation.VerifyOption
type VerifyOptions = attestation.VerifyOptions

func WithMeasurement(measurement []byte) VerifyOption {
	return attestation.WithMeasurement(measurement)
}

func WithTimestamp(timestamp time.Time) VerifyOption {
	return attestation.WithTimestamp(timestamp)
}
