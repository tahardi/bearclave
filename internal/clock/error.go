package clock

import (
	"errors"
	"fmt"
)

var (
	ErrCPU                = errors.New("cpu")
	ErrCPUTSC             = fmt.Errorf("%w: tsc", ErrCPU)
	ErrCPUTSCNotInvariant = fmt.Errorf("%w: not invariant", ErrCPUTSC)
	ErrCPUTSCFrequency    = fmt.Errorf("%w: could not determine frequency", ErrCPUTSC)
	ErrCPUVendor          = fmt.Errorf("%w: vendor not supported", ErrCPU)
	ErrTimer              = errors.New("timer")
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

func cpuErrorTSCNotInvariant(msg string, err error) error {
	return wrapError(ErrCPUTSCNotInvariant, msg, err)
}

func cpuErrorTSCFrequency(msg string, err error) error {
	return wrapError(ErrCPUTSCFrequency, msg, err)
}

func cpuErrorVendor(msg string, err error) error {
	return wrapError(ErrCPUVendor, msg, err)
}

func timerError(msg string, err error) error {
	return wrapError(ErrTimer, msg, err)
}
