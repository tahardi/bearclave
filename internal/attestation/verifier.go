package attestation

import (
	"fmt"
	"time"

	"github.com/tahardi/bearclave/internal/setup"
)

type Verifier interface {
	Verify(report []byte, options ...VerifyOption) (userdata []byte, err error)
}

func NewVerifier(platform setup.Platform) (Verifier, error) {
	switch platform {
	case setup.Nitro:
		return NewNitroVerifier()
	case setup.SEV:
		return NewSEVVerifier()
	case setup.TDX:
		return NewTDXVerifier()
	case setup.NoTEE:
		return NewNoTEEVerifier()
	default:
		return nil, fmt.Errorf("unsupported platform '%s'", platform)
	}
}

type VerifyOption func(*VerifyOptions)
type VerifyOptions struct {
	measurement string
	timestamp   time.Time
}

func WithMeasurement(measurement string) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.measurement = measurement
	}
}

func WithTimestamp(timestamp time.Time) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.timestamp = timestamp
	}
}
