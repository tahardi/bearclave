package tee_test

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/tee"
)

func TestSealHPKERequest(t *testing.T) {
	key, err := tee.NewHPKEKey()
	require.NoError(t, err)

	tests := []struct {
		name      string
		publicKey []byte
		wantErr   error
	}{
		{
			name:      "happy path",
			publicKey: key.PublicKey(),
		},
		{
			name:      "error - invalid public key",
			publicKey: make([]byte, 10),
			wantErr:   tee.ErrHPKE,
		},
		{
			name:      "error - x25519 sized key",
			publicKey: randomBytes(t, 32),
			wantErr:   tee.ErrHPKE,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			plaintext := []byte("secret input")

			// when
			msg, responseKey, err := tee.SealHPKERequest(tt.publicKey, plaintext, []byte("POST /echo"))

			// then
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, msg.Enc, 1120)
			assert.NotEqual(t, plaintext, msg.Ciphertext)
			assert.NotNil(t, responseKey)
		})
	}
}

func randomBytes(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return b
}
