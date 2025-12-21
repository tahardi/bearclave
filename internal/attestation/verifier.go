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
	Debug       bool
	Measurement string
	Nonce       []byte
	Timestamp   time.Time
}

func MakeDefaultVerifyOptions() VerifyOptions {
	return VerifyOptions{
		Debug:       false,
		Measurement: "",
		Nonce:       nil,
		Timestamp:   time.Now(),
	}
}

func WithVerifyDebug(debug bool) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.Debug = debug
	}
}

func WithVerifyMeasurement(measurement string) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.Measurement = measurement
	}
}

func WithVerifyVerifyNonce(nonce []byte) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.Nonce = nonce
	}
}

func WithVerifyTimestamp(timestamp time.Time) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.Timestamp = timestamp
	}
}
