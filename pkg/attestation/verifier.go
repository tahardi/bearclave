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

func WithDebug(debug bool) VerifyOption {
	return attestation.WithDebug(debug)
}

func WithMeasurement(measurement string) VerifyOption {
	return attestation.WithMeasurement(measurement)
}

func WithTimestamp(timestamp time.Time) VerifyOption {
	return attestation.WithTimestamp(timestamp)
}
