package bearclave

import (
	"fmt"
	"time"

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
