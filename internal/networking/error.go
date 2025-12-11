package networking

import (
	"errors"
	"fmt"
)

var (
	ErrDialer           = errors.New("dialer")
	ErrListener         = errors.New("listener")
	ErrSocketParseAddr  = errors.New("socket parse addr")
	ErrVSocketParseAddr = errors.New("vsocket parse addr")
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

func dialerError(msg string, err error) error {
	return wrapError(ErrDialer, msg, err)
}

func listenerError(msg string, err error) error {
	return wrapError(ErrListener, msg, err)
}
