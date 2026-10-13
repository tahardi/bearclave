package tee_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/tee"
)

func TestAttestKeyBinding(t *testing.T) {
	tests := []struct {
		name    string
		binding tee.KeyBinding
		wantErr error
	}{
		{
			name: "happy path",
			binding: tee.KeyBinding{
				Purpose:   tee.HPKEKeyBindingPurpose,
				PublicKey: []byte("test public key"),
				Nonce:     []byte("nonce"),
				IssuedAt:  time.Now().UTC(),
			},
		},
		{
			name: "error - missing purpose",
			binding: tee.KeyBinding{
				PublicKey: []byte("test public key"),
				IssuedAt:  time.Now(),
			},
			wantErr: tee.ErrKeyBinding,
		},
		{
			name: "error - missing public key",
			binding: tee.KeyBinding{
				Purpose:  tee.HPKEKeyBindingPurpose,
				IssuedAt: time.Now(),
			},
			wantErr: tee.ErrKeyBinding,
		},
		{
			name: "error - missing issued at",
			binding: tee.KeyBinding{
				Purpose:   tee.HPKEKeyBindingPurpose,
				PublicKey: []byte("test public key"),
			},
			wantErr: tee.ErrKeyBinding,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			attester, err := tee.NewAttester(tee.NoTEE)
			require.NoError(t, err)

			// when
			got, err := tee.AttestKeyBinding(attester, tt.binding)

			// then
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got.Base)

			var binding tee.KeyBinding
			require.NoError(t, json.Unmarshal(got.UserData, &binding))
			assert.Equal(t, tt.binding.Purpose, binding.Purpose)
			assert.Equal(t, tt.binding.PublicKey, binding.PublicKey)
			assert.Equal(t, tt.binding.Nonce, binding.Nonce)
			assert.True(t, tt.binding.IssuedAt.Equal(binding.IssuedAt))
		})
	}
}

func TestVerifyKeyBinding(t *testing.T) {
	attester, err := tee.NewAttester(tee.NoTEE)
	require.NoError(t, err)
	verifier, err := tee.NewVerifier(tee.NoTEE)
	require.NoError(t, err)

	nonce := []byte("client nonce")
	binding := tee.KeyBinding{
		Purpose:   tee.HPKEKeyBindingPurpose,
		PublicKey: []byte("test public key"),
		Nonce:     nonce,
		IssuedAt:  time.Now(),
	}
	policy := tee.KeyBindingPolicy{
		Measurement: tee.NoTEEMeasurement,
		Nonce:       nonce,
		MaxAge:      time.Hour,
	}

	withBinding := func(modify func(*tee.KeyBinding)) tee.KeyBinding {
		b := binding
		modify(&b)
		return b
	}
	withPolicy := func(modify func(*tee.KeyBindingPolicy)) tee.KeyBindingPolicy {
		p := policy
		modify(&p)
		return p
	}

	tests := []struct {
		name    string
		purpose string
		binding tee.KeyBinding
		policy  tee.KeyBindingPolicy
		wantErr error
	}{
		{
			name:    "happy path",
			purpose: tee.HPKEKeyBindingPurpose,
			binding: binding,
			policy:  policy,
		},
		{
			name:    "happy path - no nonce policy",
			purpose: tee.HPKEKeyBindingPurpose,
			binding: withBinding(func(b *tee.KeyBinding) { b.Nonce = nil }),
			policy:  withPolicy(func(p *tee.KeyBindingPolicy) { p.Nonce = nil }),
		},
		{
			name:    "error - missing measurement policy",
			purpose: tee.HPKEKeyBindingPurpose,
			binding: binding,
			policy:  withPolicy(func(p *tee.KeyBindingPolicy) { p.Measurement = "" }),
			wantErr: tee.ErrKeyBinding,
		},
		{
			name:    "error - missing max age policy",
			purpose: tee.HPKEKeyBindingPurpose,
			binding: binding,
			policy:  withPolicy(func(p *tee.KeyBindingPolicy) { p.MaxAge = 0 }),
			wantErr: tee.ErrKeyBinding,
		},
		{
			name:    "error - wrong measurement",
			purpose: tee.HPKEKeyBindingPurpose,
			binding: binding,
			policy:  withPolicy(func(p *tee.KeyBindingPolicy) { p.Measurement = "wrong" }),
			wantErr: tee.ErrVerifierMeasurement,
		},
		{
			name:    "error - wrong purpose",
			purpose: tee.HPKEKeyBindingPurpose,
			binding: withBinding(func(b *tee.KeyBinding) { b.Purpose = tee.TLSKeyBindingPurpose }),
			policy:  policy,
			wantErr: tee.ErrKeyBindingPurpose,
		},
		{
			name:    "error - wrong nonce",
			purpose: tee.HPKEKeyBindingPurpose,
			binding: withBinding(func(b *tee.KeyBinding) { b.Nonce = []byte("other client nonce") }),
			policy:  policy,
			wantErr: tee.ErrKeyBindingNonce,
		},
		{
			name:    "error - stale issued at",
			purpose: tee.HPKEKeyBindingPurpose,
			binding: withBinding(func(b *tee.KeyBinding) { b.IssuedAt = time.Now().Add(-2 * time.Hour) }),
			policy:  policy,
			wantErr: tee.ErrKeyBindingStale,
		},
		{
			name:    "error - issued in the future",
			purpose: tee.HPKEKeyBindingPurpose,
			binding: withBinding(func(b *tee.KeyBinding) { b.IssuedAt = time.Now().Add(time.Hour) }),
			policy:  policy,
			wantErr: tee.ErrKeyBindingStale,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			attestResult, err := tee.AttestKeyBinding(attester, tt.binding)
			require.NoError(t, err)

			// when
			got, err := tee.VerifyKeyBinding(verifier, attestResult, tt.purpose, tt.policy)

			// then
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.binding.Purpose, got.Purpose)
			assert.Equal(t, tt.binding.PublicKey, got.PublicKey)
			assert.Equal(t, tt.binding.Nonce, got.Nonce)
			assert.True(t, tt.binding.IssuedAt.Equal(got.IssuedAt))
		})
	}

	t.Run("error - tampered key binding", func(t *testing.T) {
		// given
		attestResult, err := tee.AttestKeyBinding(attester, binding)
		require.NoError(t, err)
		tampered := withBinding(func(b *tee.KeyBinding) { b.PublicKey = []byte("attacker public key") })
		attestResult.UserData, err = json.Marshal(tampered)
		require.NoError(t, err)

		// when
		_, err = tee.VerifyKeyBinding(verifier, attestResult, tee.HPKEKeyBindingPurpose, policy)

		// then
		require.ErrorIs(t, err, tee.ErrVerifier)
	})

	t.Run("error - tampered report", func(t *testing.T) {
		// given
		attestResult, err := tee.AttestKeyBinding(attester, binding)
		require.NoError(t, err)
		var report map[string]any
		require.NoError(t, json.Unmarshal(attestResult.Base.Report, &report))
		report["timestamp"] = time.Now().Add(-time.Minute).Unix()
		attestResult.Base.Report, err = json.Marshal(report)
		require.NoError(t, err)

		// when
		_, err = tee.VerifyKeyBinding(verifier, attestResult, tee.HPKEKeyBindingPurpose, policy)

		// then
		require.ErrorIs(t, err, tee.ErrVerifier)
	})

	t.Run("error - wrong platform", func(t *testing.T) {
		// given
		nitroVerifier, err := tee.NewVerifier(tee.Nitro)
		require.NoError(t, err)
		attestResult, err := tee.AttestKeyBinding(attester, binding)
		require.NoError(t, err)

		// when
		_, err = tee.VerifyKeyBinding(nitroVerifier, attestResult, tee.HPKEKeyBindingPurpose, policy)

		// then
		require.ErrorIs(t, err, tee.ErrVerifier)
	})

	t.Run("error - unknown field", func(t *testing.T) {
		// given
		issuedAt, err := json.Marshal(time.Now())
		require.NoError(t, err)
		userData := `{"purpose":"bearclave/hpke/v1","public_key":"AA==","issued_at":` +
			string(issuedAt) + `,"extra":1}`
		attestResult, err := attester.Attest(tee.WithAttestUserData([]byte(userData)))
		require.NoError(t, err)
		noNoncePolicy := withPolicy(func(p *tee.KeyBindingPolicy) { p.Nonce = nil })

		// when
		_, err = tee.VerifyKeyBinding(verifier, attestResult, tee.HPKEKeyBindingPurpose, noNoncePolicy)

		// then
		require.ErrorIs(t, err, tee.ErrKeyBinding)
	})
}
