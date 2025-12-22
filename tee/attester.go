package tee

import (
	"crypto/sha256"

	"github.com/tahardi/bearclave"
)

type Attester struct {
	base bearclave.Attester
}

func NewAttester(platform Platform) (*Attester, error) {
	var base bearclave.Attester
	var err error

	switch platform {
	case Nitro:
		base, err = bearclave.NewNitroAttester()
	case SEV:
		base, err = bearclave.NewSEVAttester()
	case TDX:
		base, err = bearclave.NewTDXAttester()
	case NoTEE:
		base, err = bearclave.NewNoTEEAttester()
	default:
		return nil, unsupportedPlatformError(string(platform), nil)
	}

	if err != nil {
		return nil, attesterError("making attester", err)
	}
	return NewAttesterWithBase(base)
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
		measurement, err := MeasureUserData(opts.UserData)
		if err != nil {
			return nil, attesterError("measuring output", err)
		}
		attestResult.UserData = opts.UserData
		opts.Base = append(opts.Base, bearclave.WithAttestUserData(measurement))
	}

	baseResult, err := a.base.Attest(opts.Base...)
	if err != nil {
		return nil, attesterError("base attesting", err)
	}

	attestResult.Base = baseResult
	return attestResult, nil
}

type AttestResult struct {
	Base     *bearclave.AttestResult `json:"base,omitempty"`
	UserData []byte                  `json:"userdata,omitempty"`
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

func MeasureUserData(output []byte) ([]byte, error) {
	if output == nil {
		return nil, nil
	}
	hash := sha256.Sum256(output)
	return hash[:], nil
}
