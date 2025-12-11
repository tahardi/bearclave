package attestation

import (
	"errors"
	"fmt"
)

var (
	ErrAttester                = errors.New("attester")
	ErrAttesterUserDataTooLong = fmt.Errorf("%w: user data too long", ErrAttester)
	ErrVerifier                = errors.New("verifier")
	ErrVerifierDebugMode       = fmt.Errorf("%w: debug mode", ErrVerifier)
	ErrVerifierMeasurement     = fmt.Errorf("%w: measurement", ErrVerifier)
	ErrVerifierNonce           = fmt.Errorf("%w: nonce", ErrVerifier)
	ErrVerifierTimestamp       = fmt.Errorf("%w: timestamp", ErrVerifier)
)

func wrapError(baseErr error, msg string, err error) error {
	switch {
	case msg == "" && err == nil:
		return baseErr
	case msg != "" && err != nil:
		return fmt.Errorf("%w: %s: %w", baseErr, msg, err)
	case msg != "":
		return fmt.Errorf("%w: %s", baseErr, msg)
	default:
		return fmt.Errorf("%w: %w", baseErr, err)
	}
}

func attesterError(msg string, err error) error {
	return wrapError(ErrAttester, msg, err)
}

func attesterErrorUserDataTooLong(msg string, err error) error {
	return wrapError(ErrAttesterUserDataTooLong, msg, err)
}

func verifierError(msg string, err error) error {
	return wrapError(ErrVerifier, msg, err)
}

func verifierErrorDebugMode(msg string, err error) error {
	return wrapError(ErrVerifierDebugMode, msg, err)
}

func verifierErrorMeasurement(msg string, err error) error {
	return wrapError(ErrVerifierMeasurement, msg, err)
}

func verifierErrorNonce(msg string, err error) error {
	return wrapError(ErrVerifierNonce, msg, err)
}

func verifierErrorTimestamp(msg string, err error) error {
	return wrapError(ErrVerifierTimestamp, msg, err)
}
