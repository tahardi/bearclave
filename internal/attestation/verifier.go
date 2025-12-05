package attestation

import (
	"time"
)

type Verifier interface {
	Verify(report []byte, options ...VerifyOption) (userdata []byte, err error)
}

type VerifyOption func(*VerifyOptions)
type VerifyOptions struct {
	debug       bool
	measurement string
	timestamp   time.Time
}

func WithDebug(debug bool) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.debug = debug
	}
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
