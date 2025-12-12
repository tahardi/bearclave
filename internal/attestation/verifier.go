package attestation

import (
	"time"
)

type Verifier interface {
	Verify(report *AttestResult, options ...VerifyOption) (result *VerifyResult, err error)
}

type VerifyResult struct {
	UserData  []byte `json:"userdata"`
	PublicKey []byte `json:"publickey"`
}

type VerifyOption func(*VerifyOptions)
type VerifyOptions struct {
	debug       bool
	measurement string
	nonce       []byte
	timestamp   time.Time
}

func WithVerifyDebug(debug bool) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.debug = debug
	}
}

func WithVerifyMeasurement(measurement string) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.measurement = measurement
	}
}

func WithVerifyVerifyNonce(nonce []byte) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.nonce = nonce
	}
}

func WithVerifyTimestamp(timestamp time.Time) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.timestamp = timestamp
	}
}
