package tee

import (
	"crypto/sha256"
	"encoding/json"

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

	if opts.Output != nil {
		outputMeasurement, err := CalculateOutputMeasurement(opts.Output)
		if err != nil {
			return nil, err
		}
		opts.Base = append(opts.Base, bearclave.WithAttestUserData(outputMeasurement))
	}

	baseResult, err := a.base.Attest(opts.Base...)
	if err != nil {
		return nil, teeError("", err)
	}

	attestResult := &AttestResult{Base: baseResult}
	if opts.Output != nil {
		attestResult.Output = opts.Output
	}
	return attestResult, nil
}

type AttestResult struct {
	Base   *bearclave.AttestResult `json:"base,omitempty"`
	Output json.RawMessage         `json:"output,omitempty"`
}
type AttestOption func(*AttestOptions)
type AttestOptions struct {
	Base   []bearclave.AttestOption `json:"base,omitempty"`
	Output json.RawMessage          `json:"output,omitempty"`
}

func MakeDefaultAttestOptions() AttestOptions {
	return AttestOptions{
		Base:   []bearclave.AttestOption{},
		Output: nil,
	}
}

func WithAttestNonce(nonce []byte) AttestOption {
	return func(opts *AttestOptions) {
		opts.Base = append(opts.Base, bearclave.WithAttestNonce(nonce))
	}
}

func WithAttestOutput(output json.RawMessage) AttestOption {
	return func(opts *AttestOptions) {
		opts.Output = output
	}
}

func CalculateOutputMeasurement(output json.RawMessage) ([]byte, error) {
	if output == nil {
		return nil, nil
	}
	hash := sha256.Sum256(output)
	return hash[:], nil
}
