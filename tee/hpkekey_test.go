package tee_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/tee"
)

func TestHPKEKey_Open(t *testing.T) {
	plaintext := []byte("secret input")
	aad := []byte("POST /echo")
	tests := []struct {
		name    string
		mutate  func(t *testing.T, key *tee.HPKEKey, msg *tee.SealedMessage, aad *[]byte) *tee.HPKEKey
		wantErr error
	}{
		{
			name: "happy path",
			mutate: func(_ *testing.T, key *tee.HPKEKey, _ *tee.SealedMessage, _ *[]byte) *tee.HPKEKey {
				return key
			},
		},
		{
			name: "error - tampered ciphertext",
			mutate: func(_ *testing.T, key *tee.HPKEKey, msg *tee.SealedMessage, _ *[]byte) *tee.HPKEKey {
				msg.Ciphertext[len(msg.Ciphertext)-1] ^= 0xff
				return key
			},
			wantErr: tee.ErrHPKE,
		},
		{
			name: "error - tampered enc",
			mutate: func(_ *testing.T, key *tee.HPKEKey, msg *tee.SealedMessage, _ *[]byte) *tee.HPKEKey {
				msg.Enc[0] ^= 0xff
				return key
			},
			wantErr: tee.ErrHPKE,
		},
		{
			name: "error - wrong aad",
			mutate: func(_ *testing.T, key *tee.HPKEKey, _ *tee.SealedMessage, aad *[]byte) *tee.HPKEKey {
				*aad = []byte("POST /other")
				return key
			},
			wantErr: tee.ErrHPKE,
		},
		{
			name: "error - wrong key",
			mutate: func(t *testing.T, _ *tee.HPKEKey, _ *tee.SealedMessage, _ *[]byte) *tee.HPKEKey {
				t.Helper()
				other, err := tee.NewHPKEKey()
				require.NoError(t, err)
				return other
			},
			wantErr: tee.ErrHPKE,
		},
		{
			name: "error - missing enc",
			mutate: func(_ *testing.T, key *tee.HPKEKey, msg *tee.SealedMessage, _ *[]byte) *tee.HPKEKey {
				msg.Enc = nil
				return key
			},
			wantErr: tee.ErrHPKE,
		},
		{
			name: "error - truncated ciphertext",
			mutate: func(_ *testing.T, key *tee.HPKEKey, msg *tee.SealedMessage, _ *[]byte) *tee.HPKEKey {
				msg.Ciphertext = msg.Ciphertext[:5]
				return key
			},
			wantErr: tee.ErrHPKE,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			key, err := tee.NewHPKEKey()
			require.NoError(t, err)

			msg, clientResponseKey, err := tee.SealHPKERequest(key.PublicKey(), plaintext, aad)
			require.NoError(t, err)

			openAAD := aad
			openKey := tt.mutate(t, key, &msg, &openAAD)

			// when
			got, serverResponseKey, err := openKey.Open(msg, openAAD)

			// then
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, plaintext, got)

			reply, err := clientResponseKey.Open(serverResponseKey.Seal([]byte("reply")))
			require.NoError(t, err)
			assert.Equal(t, []byte("reply"), reply)
		})
	}
}
