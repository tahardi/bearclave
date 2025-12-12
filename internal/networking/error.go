package networking

import (
	"errors"
	"fmt"
)

var (
	ErrDialContext     = errors.New("dial context")
	ErrListener        = errors.New("listener")
	ErrSocketParseAddr = errors.New("socket parse addr")
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

func dialContextError(msg string, err error) error {
	return wrapError(ErrDialContext, msg, err)
}

func listenerError(msg string, err error) error {
	return wrapError(ErrListener, msg, err)
}
