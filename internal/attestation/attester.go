package attestation

import (
	"errors"
	"fmt"
)

var (
	ErrAttester = errors.New("attester")
	ErrAttesterUserDataTooLong = fmt.Errorf("%w: user data too long", ErrAttester)
)

type Attester interface {
	Attest(options ...AttestOption) (result *AttestResult, err error)
}

type AttestResult struct {
	Report []byte `json:"report"`
}

type AttestOption func(*AttestOptions)
type AttestOptions struct {
	nonce     []byte
	publicKey []byte
	userData  []byte
}

func WithAttestNonce(nonce []byte) AttestOption {
	return func(opts *AttestOptions) {
		opts.nonce = nonce
	}
}

func WithPublicKey(publicKey []byte) AttestOption {
	return func(opts *AttestOptions) {
		opts.publicKey = publicKey
	}
}

func WithUserData(userData []byte) AttestOption {
	return func(opts *AttestOptions) {
		opts.userData = userData
	}
}

func wrapAttesterError(attesterErr error, msg string, err error) error {
	switch {
	case msg == "" && err == nil:
		return attesterErr
	case msg != "" && err != nil:
		return fmt.Errorf("%w: %s: %w", attesterErr, msg, err)
	case msg != "":
		return fmt.Errorf("%w: %s", attesterErr, msg)
	default:
		return fmt.Errorf("%w: %w", attesterErr, err)
	}
}

func attesterError(msg string, err error) error {
	return wrapAttesterError(ErrAttester, msg, err)
}

func attesterErrorUserDataTooLong(msg string, err error) error {
	return wrapAttesterError(ErrAttesterUserDataTooLong, msg, err)
}
