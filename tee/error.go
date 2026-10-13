package tee

import (
	"errors"
	"fmt"

	"github.com/tahardi/bearclave"
)

var (
	ErrAttester            = bearclave.ErrAttester
	ErrAttesterUserData    = bearclave.ErrAttesterUserData
	ErrDialContext         = bearclave.ErrDialContext
	ErrListener            = bearclave.ErrListener
	ErrProxy               = errors.New("proxy")
	ErrReverseProxy        = errors.New("reverse proxy")
	ErrServer              = errors.New("server")
	ErrSocket              = errors.New("socket")
	ErrTimer               = bearclave.ErrTimer
	ErrVerifier            = bearclave.ErrVerifier
	ErrVerifierDebugMode   = bearclave.ErrVerifierDebugMode
	ErrVerifierMeasurement = bearclave.ErrVerifierMeasurement
	ErrVerifierNonce       = bearclave.ErrVerifierNonce
	ErrVerifierTimestamp   = bearclave.ErrVerifierTimestamp
	ErrCertProvider        = errors.New("cert provider")
	ErrKeyBinding          = errors.New("key binding")
	ErrKeyBindingPurpose   = fmt.Errorf("%w: purpose", ErrKeyBinding)
	ErrKeyBindingNonce     = fmt.Errorf("%w: nonce", ErrKeyBinding)
	ErrKeyBindingStale     = fmt.Errorf("%w: stale", ErrKeyBinding)
	ErrKeyBindingPublicKey = fmt.Errorf("%w: public key", ErrKeyBinding)
	ErrUnsupportedPlatform = errors.New("unsupported platform")
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

func certProviderError(msg string, err error) error {
	return wrapError(ErrCertProvider, msg, err)
}

func keyBindingError(msg string, err error) error {
	return wrapError(ErrKeyBinding, msg, err)
}

func keyBindingPurposeError(msg string, err error) error {
	return wrapError(ErrKeyBindingPurpose, msg, err)
}

func keyBindingNonceError(msg string, err error) error {
	return wrapError(ErrKeyBindingNonce, msg, err)
}

func keyBindingStaleError(msg string, err error) error {
	return wrapError(ErrKeyBindingStale, msg, err)
}

func keyBindingPublicKeyError(msg string, err error) error {
	return wrapError(ErrKeyBindingPublicKey, msg, err)
}

func proxyError(msg string, err error) error {
	return wrapError(ErrProxy, msg, err)
}

func reverseProxyError(msg string, err error) error {
	return wrapError(ErrReverseProxy, msg, err)
}

func serverError(msg string, err error) error {
	return wrapError(ErrServer, msg, err)
}

func socketError(msg string, err error) error {
	return wrapError(ErrSocket, msg, err)
}

func verifierError(msg string, err error) error {
	return wrapError(ErrVerifier, msg, err)
}

func unsupportedPlatformError(msg string, err error) error {
	return wrapError(ErrUnsupportedPlatform, msg, err)
}
