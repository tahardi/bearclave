package tee

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/tahardi/bearclave"
)

type Verifier struct {
	base bearclave.Verifier
}

func NewVerifier(platform Platform) (*Verifier, error) {
	var base bearclave.Verifier
	var err error

	switch platform {
	case Nitro:
		base, err = bearclave.NewNitroVerifier()
	case SEV:
		base, err = bearclave.NewSEVVerifier()
	case TDX:
		base, err = bearclave.NewTDXVerifier()
	case NoTEE:
		base, err = bearclave.NewNoTEEVerifier()
	default:
		return nil, unsupportedPlatformError(string(platform), nil)
	}

	if err != nil {
		return nil, verifierError("making verifier", err)
	}
	return NewVerifierWithBase(base)
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
		return nil, verifierError("verifying base attestResult", err)
	case attestResult.UserData == nil && len(baseResult.UserData) != 0:
		return nil, verifierError("missing output", nil)
	case attestResult.UserData != nil && len(baseResult.UserData) == 0:
		return nil, verifierError("missing user data", nil)
	}

	verifyResult := &VerifyResult{Base: baseResult, UserData: attestResult.UserData}
	if len(baseResult.UserData) == 0 {
		return verifyResult, nil
	}

	err = VerifyUserData(baseResult.UserData, verifyResult.UserData)
	if err != nil {
		return nil, err
	}
	return verifyResult, nil
}

type VerifyResult struct {
	Base     *bearclave.VerifyResult `json:"base"`
	UserData []byte                  `json:"userdata,omitempty"`
}
type VerifyOption func(*VerifyOptions)
type VerifyOptions struct {
	Base []bearclave.VerifyOption `json:"base,omitempty"`
}

func MakeDefaultVerifyOptions() VerifyOptions {
	return VerifyOptions{
		Base: []bearclave.VerifyOption{},
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

func WithVerifyTimestamp(timestamp time.Time) VerifyOption {
	return func(opts *VerifyOptions) {
		opts.Base = append(opts.Base, bearclave.WithVerifyTimestamp(timestamp))
	}
}

func VerifyUserData(expectedMeasurement []byte, userData []byte) error {
	gotMeasurement, err := MeasureUserData(userData)
	if err != nil {
		return verifierError("measuring user data", err)
	}

	// The SEV and TDX TEE platforms always return 64 bytes for user data
	// even if the provided user data was shorter. Ensure that we use the
	// correct length when comparing so we don't falsely mismatch
	correctedMeasurement := expectedMeasurement[:len(gotMeasurement)]
	if !bytes.Equal(correctedMeasurement, gotMeasurement) {
		msg := fmt.Sprintf(
			"user data measurement mismatch: expected %s, got %s",
			base64.StdEncoding.EncodeToString(correctedMeasurement),
			base64.StdEncoding.EncodeToString(gotMeasurement),
		)
		return verifierError(msg, nil)
	}
	return nil
}
