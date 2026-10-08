package tee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

const (
	HPKEKeyBindingPurpose      = "bearclave/hpke/v1"
	TLSKeyBindingPurpose       = "bearclave/tls/v1"
	DefaultKeyBindingClockSkew = time.Minute
)

type KeyBinding struct {
	Purpose   string    `json:"purpose"`
	PublicKey []byte    `json:"public_key"`
	Nonce     []byte    `json:"nonce,omitempty"`
	IssuedAt  time.Time `json:"issued_at"`
}

func AttestKeyBinding(attester *Attester, binding KeyBinding) (*AttestResult, error) {
	switch {
	case binding.Purpose == "":
		return nil, keyBindingError("missing purpose", nil)
	case len(binding.PublicKey) == 0:
		return nil, keyBindingError("missing public key", nil)
	case binding.IssuedAt.IsZero():
		return nil, keyBindingError("missing issued at", nil)
	}

	data, err := json.Marshal(binding)
	if err != nil {
		return nil, keyBindingError("marshaling key binding", err)
	}

	attestResult, err := attester.Attest(WithAttestUserData(data))
	if err != nil {
		return nil, keyBindingError("attesting key binding", err)
	}
	return attestResult, nil
}

func VerifyKeyBinding(
	verifier *Verifier,
	attestResult *AttestResult,
	purpose string,
	policy KeyBindingPolicy,
) (*KeyBinding, error) {
	switch {
	case policy.Measurement == "":
		return nil, keyBindingError("missing measurement policy", nil)
	case policy.MaxAge <= 0:
		return nil, keyBindingError("missing max age policy", nil)
	}

	now := policy.Now
	if now.IsZero() {
		now = time.Now()
	}

	verified, err := verifier.Verify(
		attestResult,
		WithVerifyMeasurement(policy.Measurement),
		WithVerifyDebug(policy.Debug),
		WithVerifyTimestamp(now),
	)
	if err != nil {
		return nil, err
	}

	if len(verified.UserData) == 0 {
		return nil, keyBindingError("missing key binding", nil)
	}

	var binding KeyBinding
	decoder := json.NewDecoder(bytes.NewReader(verified.UserData))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&binding)
	if err != nil {
		return nil, keyBindingError("unmarshaling key binding", err)
	}

	switch {
	case binding.Purpose != purpose:
		msg := fmt.Sprintf("expected %q, got %q", purpose, binding.Purpose)
		return nil, keyBindingPurposeError(msg, nil)
	case len(binding.PublicKey) == 0:
		return nil, keyBindingPublicKeyError("missing public key", nil)
	case policy.Nonce != nil && !bytes.Equal(policy.Nonce, binding.Nonce):
		return nil, keyBindingNonceError("mismatched nonce", nil)
	case binding.IssuedAt.IsZero():
		return nil, keyBindingStaleError("missing issued at", nil)
	case binding.IssuedAt.After(now.Add(DefaultKeyBindingClockSkew)):
		return nil, keyBindingStaleError("issued in the future", nil)
	}

	age := now.Sub(binding.IssuedAt)
	if age > policy.MaxAge {
		msg := fmt.Sprintf("age %s exceeds max age %s", age, policy.MaxAge)
		return nil, keyBindingStaleError(msg, nil)
	}
	return &binding, nil
}
