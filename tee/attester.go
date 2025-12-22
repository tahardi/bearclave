package tee

import (
	"crypto/sha256"

	"github.com/tahardi/bearclave"
)

type Attester struct {
	base bearclave.Attester
}

func NewAttester(platform Platform) (*Attester, error) {
	switch platform {
	case Nitro:
		base, err := bearclave.NewNitroAttester()
		if err != nil {
			return nil, teeError("", err)
		}
		return NewAttesterWithBase(base)
	case SEV:
		base, err := bearclave.NewSEVAttester()
		if err != nil {
			return nil, teeError("", err)
		}
		return NewAttesterWithBase(base)
	case TDX:
		base, err := bearclave.NewTDXAttester()
		if err != nil {
			return nil, teeError("", err)
		}
		return NewAttesterWithBase(base)
	case NoTEE:
		base, err := bearclave.NewNoTEEAttester()
		if err != nil {
			return nil, teeError("", err)
		}
		return NewAttesterWithBase(base)
	default:
		return nil, teeErrorUnsupportedPlatform(string(platform), nil)
	}
}

func NewAttesterWithBase(base bearclave.Attester) (*Attester, error) {
	return &Attester{base: base}, nil
}

func (a *Attester) Attest(options ...AttestOption) (*AttestResult, error) {
	opts := MakeDefaultAttestOptions()
	for _, opt := range options {
		opt(&opts)
	}

	attestResult := &AttestResult{}
	if opts.UserData != nil {
		measurement, err := MeasureOutput(opts.UserData)
		if err != nil {
			return nil, err
		}
		attestResult.Output = opts.UserData
		opts.Base = append(opts.Base, bearclave.WithAttestUserData(measurement))
	}

	baseResult, err := a.base.Attest(opts.Base...)
	if err != nil {
		return nil, teeError("", err)
	}

	attestResult.Base = baseResult
	return attestResult, nil
}

type AttestResult struct {
	Base   *bearclave.AttestResult `json:"base,omitempty"`
	Output []byte                  `json:"output,omitempty"`
}
type AttestOption func(*AttestOptions)
type AttestOptions struct {
	Base     []bearclave.AttestOption `json:"base,omitempty"`
	UserData []byte                   `json:"output,omitempty"`
}

func MakeDefaultAttestOptions() AttestOptions {
	return AttestOptions{
		Base:     []bearclave.AttestOption{},
		UserData: nil,
	}
}

func WithAttestNonce(nonce []byte) AttestOption {
	return func(opts *AttestOptions) {
		opts.Base = append(opts.Base, bearclave.WithAttestNonce(nonce))
	}
}

func WithAttestUserData(userData []byte) AttestOption {
	return func(opts *AttestOptions) {
		opts.UserData = userData
	}
}

func MeasureOutput(output []byte) ([]byte, error) {
	if output == nil {
		return nil, nil
	}
	hash := sha256.Sum256(output)
	return hash[:], nil
}
