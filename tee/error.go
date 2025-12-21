package tee

import (
	"errors"
	"fmt"

	"github.com/tahardi/bearclave"
)

var (
	ErrTEE = errors.New("tee")
	ErrTEEUnsupportedPlatform = fmt.Errorf("%w: unsupported platform", ErrTEE)
	ErrTEEVerifier = teeError("", bearclave.ErrVerifier)
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

func teeError(msg string, err error) error {
	return wrapError(ErrTEE, msg, err)
}

func teeErrorUnsupportedPlatform(msg string, err error) error {
	return wrapError(ErrTEEUnsupportedPlatform, msg, err)
}

func teeErrorVerifier(msg string, err error) error {
	return wrapError(ErrTEEVerifier, msg, err)
}