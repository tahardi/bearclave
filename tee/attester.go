package tee

import (
	"fmt"

	"github.com/tahardi/bearclave"
)

var (
	ErrAttester = bearclave.ErrAttester
	ErrAttesterUserDataTooLong = bearclave.ErrAttesterUserDataTooLong
)

type Attester = bearclave.Attester

func NewAttester(platform Platform) (Attester, error) {
	switch platform {
	case Nitro:
		return bearclave.NewNitroAttester()
	case SEV:
		return bearclave.NewSEVAttester()
	case TDX:
		return bearclave.NewTDXAttester()
	case NoTEE:
		return bearclave.NewNoTEEAttester()
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, platform)
	}
}

type AttestResult = bearclave.AttestResult
type AttestOption = bearclave.AttestOption
type AttestOptions = bearclave.AttestOptions

var (
	WithAttestNonce     = bearclave.WithAttestNonce
	WithAttestUserData  = bearclave.WithAttestUserData
)
