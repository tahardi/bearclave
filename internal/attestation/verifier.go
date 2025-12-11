package attestation

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrVerifier = errors.New("verifier")
	ErrVerifierDebugMode = fmt.Errorf("%w: debug mode", ErrVerifier)
	ErrVerifierMeasurement = fmt.Errorf("%w: measurement", ErrVerifier)
	ErrVerifierNonce = fmt.Errorf("%w: nonce", ErrVerifier)
	ErrVerifierTimestamp = fmt.Errorf("%w: timestamp", ErrVerifier)
)

type Verifier interface {
	Verify(report *AttestResult, options ...VerifyOption) (result *VerifyResult, err error)
}

type VerifyResult struct {
	UserData        []byte `json:"userdata"`
	PublicKey       []byte `json:"publickey"`
}

type VerifyOption func(*VerifyOptions)
type VerifyOptions struct {
	debug       bool
	measurement string
	nonce       []byte
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

func WithVerifyNonce(nonce []byte) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.nonce = nonce
	}
}

func WithTimestamp(timestamp time.Time) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.timestamp = timestamp
	}
}

func wrapVerifierError(verifierErr error, msg string, err error) error {
	switch {
	case msg == "" && err == nil:
		return verifierErr
	case msg != "" && err != nil:
		return fmt.Errorf("%w: %s: %w", verifierErr, msg, err)
	case msg != "":
		return fmt.Errorf("%w: %s", verifierErr, msg)
	default:
		return fmt.Errorf("%w: %w", verifierErr, err)
	}
}

func verifierError(msg string, err error) error {
	return wrapVerifierError(ErrVerifier, msg, err)
}

func verifierErrorDebugMode(msg string, err error) error {
	return wrapVerifierError(ErrVerifierDebugMode, msg, err)
}

func verifierErrorMeasurement(msg string, err error) error {
	return wrapVerifierError(ErrVerifierMeasurement, msg, err)
}

func verifierErrorNonce(msg string, err error) error {
	return wrapVerifierError(ErrVerifierNonce, msg, err)
}

func verifierErrorTimestamp(msg string, err error) error {
	return wrapVerifierError(ErrVerifierTimestamp, msg, err)
}
