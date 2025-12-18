package tee

import (
	"fmt"

	"github.com/tahardi/bearclave"
)

var (
	ErrVerifier            = bearclave.ErrVerifier
	ErrVerifierDebugMode   = bearclave.ErrVerifierDebugMode
	ErrVerifierMeasurement = bearclave.ErrVerifierMeasurement
	ErrVerifierNonce       = bearclave.ErrVerifierNonce
	ErrVerifierTimestamp   = bearclave.ErrVerifierTimestamp
)

type Verifier = bearclave.Verifier

func NewVerifier(platform Platform) (Verifier, error) {
	switch platform {
	case Nitro:
		return bearclave.NewNitroVerifier()
	case SEV:
		return bearclave.NewSEVVerifier()
	case TDX:
		return bearclave.NewTDXVerifier()
	case NoTEE:
		return bearclave.NewNoTEEVerifier()
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, platform)
	}
}

type VerifyResult = bearclave.VerifyResult
type VerifyOption = bearclave.VerifyOption
type VerifyOptions = bearclave.VerifyOptions

var (
	WithVerifyDebug       = bearclave.WithVerifyDebug
	WithVerifyMeasurement = bearclave.WithVerifyMeasurement
	WithVerifyTimestamp   = bearclave.WithVerifyTimestamp
	WithVerifyNonce       = bearclave.WithVerifyNonce
)
