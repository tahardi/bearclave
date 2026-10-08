package tee_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/tee"
)

func newTestKeyBinding(
	t *testing.T,
	purpose string,
	nonce []byte,
	issuedAt time.Time,
) (*tee.AttestResult, tee.KeyBinding) {
	t.Helper()
	attester, err := tee.NewAttester(tee.NoTEE)
	require.NoError(t, err)

	binding := tee.KeyBinding{
		Purpose:   purpose,
		PublicKey: []byte("test public key"),
		Nonce:     nonce,
		IssuedAt:  issuedAt,
	}
	attestResult, err := tee.AttestKeyBinding(attester, binding)
	require.NoError(t, err)
	return attestResult, binding
}

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
	nonce := []byte("client nonce")
	defaultPolicy := tee.KeyBindingPolicy{
		Measurement: tee.NoTEEMeasurement,
		Nonce:       nonce,
		MaxAge:      time.Hour,
	}
	newVerifier := func(t *testing.T, platform tee.Platform) *tee.Verifier {
		t.Helper()
		verifier, err := tee.NewVerifier(platform)
		require.NoError(t, err)
		return verifier
	}

	type setupResult struct {
		attestResult *tee.AttestResult
		verifier     *tee.Verifier
		purpose      string
		policy       tee.KeyBindingPolicy
		want         *tee.KeyBinding
	}

	tests := []struct {
		name    string
		setup   func(t *testing.T) setupResult
		wantErr error
	}{
		{
			name: "happy path",
			setup: func(t *testing.T) setupResult {
				t.Helper()
				attestResult, binding := newTestKeyBinding(t, tee.HPKEKeyBindingPurpose, nonce, time.Now())
				return setupResult{
					attestResult: attestResult,
					verifier:     newVerifier(t, tee.NoTEE),
					purpose:      tee.HPKEKeyBindingPurpose,
					policy:       defaultPolicy,
					want:         &binding,
				}
			},
		},
		{
			name: "happy path - no nonce policy",
			setup: func(t *testing.T) setupResult {
				t.Helper()
				attestResult, binding := newTestKeyBinding(t, tee.HPKEKeyBindingPurpose, nil, time.Now())
				policy := defaultPolicy
				policy.Nonce = nil
				return setupResult{
					attestResult: attestResult,
					verifier:     newVerifier(t, tee.NoTEE),
					purpose:      tee.HPKEKeyBindingPurpose,
					policy:       policy,
					want:         &binding,
				}
			},
		},
		{
			name: "error - missing measurement policy",
			setup: func(t *testing.T) setupResult {
				t.Helper()
				attestResult, _ := newTestKeyBinding(t, tee.HPKEKeyBindingPurpose, nonce, time.Now())
				policy := defaultPolicy
				policy.Measurement = ""
				return setupResult{
					attestResult: attestResult,
					verifier:     newVerifier(t, tee.NoTEE),
					purpose:      tee.HPKEKeyBindingPurpose,
					policy:       policy,
				}
			},
			wantErr: tee.ErrKeyBinding,
		},
		{
			name: "error - missing max age policy",
			setup: func(t *testing.T) setupResult {
				t.Helper()
				attestResult, _ := newTestKeyBinding(t, tee.HPKEKeyBindingPurpose, nonce, time.Now())
				policy := defaultPolicy
				policy.MaxAge = 0
				return setupResult{
					attestResult: attestResult,
					verifier:     newVerifier(t, tee.NoTEE),
					purpose:      tee.HPKEKeyBindingPurpose,
					policy:       policy,
				}
			},
			wantErr: tee.ErrKeyBinding,
		},
		{
			name: "error - wrong measurement",
			setup: func(t *testing.T) setupResult {
				t.Helper()
				attestResult, _ := newTestKeyBinding(t, tee.HPKEKeyBindingPurpose, nonce, time.Now())
				policy := defaultPolicy
				policy.Measurement = "wrong"
				return setupResult{
					attestResult: attestResult,
					verifier:     newVerifier(t, tee.NoTEE),
					purpose:      tee.HPKEKeyBindingPurpose,
					policy:       policy,
				}
			},
			wantErr: tee.ErrVerifierMeasurement,
		},
		{
			name: "error - wrong purpose",
			setup: func(t *testing.T) setupResult {
				t.Helper()
				attestResult, _ := newTestKeyBinding(t, tee.TLSKeyBindingPurpose, nonce, time.Now())
				return setupResult{
					attestResult: attestResult,
					verifier:     newVerifier(t, tee.NoTEE),
					purpose:      tee.HPKEKeyBindingPurpose,
					policy:       defaultPolicy,
				}
			},
			wantErr: tee.ErrKeyBindingPurpose,
		},
		{
			name: "error - wrong nonce",
			setup: func(t *testing.T) setupResult {
				t.Helper()
				attestResult, _ := newTestKeyBinding(
					t,
					tee.HPKEKeyBindingPurpose,
					[]byte("other client nonce"),
					time.Now(),
				)
				return setupResult{
					attestResult: attestResult,
					verifier:     newVerifier(t, tee.NoTEE),
					purpose:      tee.HPKEKeyBindingPurpose,
					policy:       defaultPolicy,
				}
			},
			wantErr: tee.ErrKeyBindingNonce,
		},
		{
			name: "error - stale issued at",
			setup: func(t *testing.T) setupResult {
				t.Helper()
				attestResult, _ := newTestKeyBinding(
					t,
					tee.HPKEKeyBindingPurpose,
					nonce,
					time.Now().Add(-2*time.Hour),
				)
				return setupResult{
					attestResult: attestResult,
					verifier:     newVerifier(t, tee.NoTEE),
					purpose:      tee.HPKEKeyBindingPurpose,
					policy:       defaultPolicy,
				}
			},
			wantErr: tee.ErrKeyBindingStale,
		},
		{
			name: "error - issued in the future",
			setup: func(t *testing.T) setupResult {
				t.Helper()
				attestResult, _ := newTestKeyBinding(
					t,
					tee.HPKEKeyBindingPurpose,
					nonce,
					time.Now().Add(time.Hour),
				)
				return setupResult{
					attestResult: attestResult,
					verifier:     newVerifier(t, tee.NoTEE),
					purpose:      tee.HPKEKeyBindingPurpose,
					policy:       defaultPolicy,
				}
			},
			wantErr: tee.ErrKeyBindingStale,
		},
		{
			name: "error - tampered key binding",
			setup: func(t *testing.T) setupResult {
				t.Helper()
				attestResult, binding := newTestKeyBinding(t, tee.HPKEKeyBindingPurpose, nonce, time.Now())
				binding.PublicKey = []byte("attacker public key")
				tampered, err := json.Marshal(binding)
				require.NoError(t, err)
				attestResult.UserData = tampered
				return setupResult{
					attestResult: attestResult,
					verifier:     newVerifier(t, tee.NoTEE),
					purpose:      tee.HPKEKeyBindingPurpose,
					policy:       defaultPolicy,
				}
			},
			wantErr: tee.ErrVerifier,
		},
		{
			name: "error - tampered report",
			setup: func(t *testing.T) setupResult {
				t.Helper()
				attestResult, _ := newTestKeyBinding(t, tee.HPKEKeyBindingPurpose, nonce, time.Now())
				var report map[string]any
				require.NoError(t, json.Unmarshal(attestResult.Base.Report, &report))
				report["timestamp"] = time.Now().Add(-time.Minute).Unix()
				tampered, err := json.Marshal(report)
				require.NoError(t, err)
				attestResult.Base.Report = tampered
				return setupResult{
					attestResult: attestResult,
					verifier:     newVerifier(t, tee.NoTEE),
					purpose:      tee.HPKEKeyBindingPurpose,
					policy:       defaultPolicy,
				}
			},
			wantErr: tee.ErrVerifier,
		},
		{
			name: "error - wrong platform",
			setup: func(t *testing.T) setupResult {
				t.Helper()
				attestResult, _ := newTestKeyBinding(t, tee.HPKEKeyBindingPurpose, nonce, time.Now())
				return setupResult{
					attestResult: attestResult,
					verifier:     newVerifier(t, tee.Nitro),
					purpose:      tee.HPKEKeyBindingPurpose,
					policy:       defaultPolicy,
				}
			},
			wantErr: tee.ErrVerifier,
		},
		{
			name: "error - unknown field",
			setup: func(t *testing.T) setupResult {
				t.Helper()
				attester, err := tee.NewAttester(tee.NoTEE)
				require.NoError(t, err)
				issuedAt, err := json.Marshal(time.Now())
				require.NoError(t, err)
				userData := `{"purpose":"bearclave/hpke/v1","public_key":"AA==","issued_at":` +
					string(issuedAt) + `,"extra":1}`
				attestResult, err := attester.Attest(tee.WithAttestUserData([]byte(userData)))
				require.NoError(t, err)
				policy := defaultPolicy
				policy.Nonce = nil
				return setupResult{
					attestResult: attestResult,
					verifier:     newVerifier(t, tee.NoTEE),
					purpose:      tee.HPKEKeyBindingPurpose,
					policy:       policy,
				}
			},
			wantErr: tee.ErrKeyBinding,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			s := tt.setup(t)

			// when
			got, err := tee.VerifyKeyBinding(s.verifier, s.attestResult, s.purpose, s.policy)

			// then
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, s.want.Purpose, got.Purpose)
			assert.Equal(t, s.want.PublicKey, got.PublicKey)
			assert.Equal(t, s.want.Nonce, got.Nonce)
			assert.True(t, s.want.IssuedAt.Equal(got.IssuedAt))
		})
	}
}
