package tee

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tahardi/bearclave"
)

type Verifier struct {
	base bearclave.Verifier
}

func NewVerifier(platform Platform) (*Verifier, error) {
	switch platform {
	case Nitro:
		base, err := bearclave.NewNitroVerifier()
		if err != nil {
			return nil, teeError("", err)
		}
		return NewVerifierWithBase(base)
	case SEV:
		base, err := bearclave.NewSEVVerifier()
		if err != nil {
			return nil, teeError("", err)
		}
		return NewVerifierWithBase(base)
	case TDX:
		base, err := bearclave.NewTDXVerifier()
		if err != nil {
			return nil, teeError("", err)
		}
		return NewVerifierWithBase(base)
	case NoTEE:
		base, err := bearclave.NewNoTEEVerifier()
		if err != nil {
			return nil, teeError("", err)
		}
		return NewVerifierWithBase(base)
	default:
		return nil, teeErrorUnsupportedPlatform(string(platform), nil)
	}
}

func NewVerifierWithBase(base bearclave.Verifier) (*Verifier, error) {
	return &Verifier{base: base}, nil
}

func (v *Verifier) Verify(
	attestResult *AttestResult,
	options ...VerifyOption,
) (*VerifyResult, error) {
	opts := MakeDefaultVerifyOptions()
	for _, opt := range options {
		opt(&opts)
	}

	baseResult, err := v.base.Verify(attestResult.Base, opts.Base...)
	switch {
	case err != nil:
		return nil, teeError("", err)
	case attestResult.Output == nil && len(baseResult.UserData) != 0:
		return nil, teeErrorVerifier("missing output", nil)
	case attestResult.Output != nil && len(baseResult.UserData) == 0:
		return nil, teeErrorVerifier("missing user data", nil)
	}

	verifyResult := &VerifyResult{Base: baseResult, Output: attestResult.Output}
	if len(baseResult.UserData) == 0 {
		return verifyResult, nil
	}

	err = VerifyOutput(baseResult.UserData, verifyResult.Output)
	if err != nil {
		return nil, teeError("", err)
	}
	return verifyResult, nil
}

type VerifyResult struct {
	Base   *bearclave.VerifyResult `json:"base"`
	Output json.RawMessage         `json:"output,omitempty"`
}
type VerifyOption func(*VerifyOptions)
type VerifyOptions struct {
	Base   []bearclave.VerifyOption `json:"base,omitempty"`
	Output json.RawMessage          `json:"output,omitempty"`
}

func MakeDefaultVerifyOptions() VerifyOptions {
	return VerifyOptions{
		Base:   []bearclave.VerifyOption{},
		Output: nil,
	}
}

func WithVerifyDebug(debug bool) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.Base = append(opts.Base, bearclave.WithVerifyDebug(debug))
	}
}

func WithVerifyMeasurement(measurement string) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.Base = append(opts.Base, bearclave.WithVerifyMeasurement(measurement))
	}
}

func WithVerifyNonce(nonce []byte) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.Base = append(opts.Base, bearclave.WithVerifyNonce(nonce))
	}
}

func WithVerifyOutput(output json.RawMessage) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.Output = output
	}
}

func WithVerifyTimestamp(timestamp time.Time) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.Base = append(opts.Base, bearclave.WithVerifyTimestamp(timestamp))
	}
}

func VerifyOutput(expectedMeasurement []byte, output json.RawMessage) error {
	outputMeasurement, err := CalculateOutputMeasurement(output)
	if err != nil {
		return err
	}

	// The SEV and TDX TEE platforms always return 64 bytes for user data
	// even if the provided user data was shorter. Ensure that we use the
	// correct length when comparing so we don't falsely mismatch
	correctedMeasurement := expectedMeasurement[:len(outputMeasurement)]
	if !bytes.Equal(correctedMeasurement, outputMeasurement) {
		msg := fmt.Sprintf(
			"output measurement mismatch: expected %s, got %s",
			base64.StdEncoding.EncodeToString(correctedMeasurement),
			base64.StdEncoding.EncodeToString(outputMeasurement),
		)
		return teeErrorVerifier(msg, nil)
	}
	return nil
}
