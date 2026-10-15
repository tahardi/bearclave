package tee_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/tee"
)

func TestNewHPKEResponseKey(t *testing.T) {
	tests := []struct {
		name    string
		secret  []byte
		wantErr error
	}{
		{
			name:   "happy path",
			secret: bytes.Repeat([]byte{1}, tee.HPKEResponseSecretSize),
		},
		{
			name:    "error - short secret",
			secret:  bytes.Repeat([]byte{1}, 16),
			wantErr: tee.ErrHPKE,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			secret := tt.secret

			// when
			key, err := tee.NewHPKEResponseKey(secret)

			// then
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, key)
		})
	}
}

func TestHPKEResponseKey_Seal(t *testing.T) {
	t.Run("happy path - random nonce", func(t *testing.T) {
		// given
		key, err := tee.NewHPKEResponseKey(bytes.Repeat([]byte{1}, tee.HPKEResponseSecretSize))
		require.NoError(t, err)
		plaintext := []byte("reply")

		// when
		first := key.Seal(plaintext)
		second := key.Seal(plaintext)

		// then
		assert.NotEqual(t, first.Ciphertext, second.Ciphertext)
		got, err := key.Open(first)
		require.NoError(t, err)
		assert.Equal(t, plaintext, got)
		got, err = key.Open(second)
		require.NoError(t, err)
		assert.Equal(t, plaintext, got)
	})
}

func TestHPKEResponseKey_Open(t *testing.T) {
	secret := bytes.Repeat([]byte{1}, tee.HPKEResponseSecretSize)
	tests := []struct {
		name       string
		openSecret []byte
		mutate     func(msg *tee.SealedMessage)
		wantErr    error
	}{
		{
			name:       "happy path",
			openSecret: secret,
			mutate:     func(_ *tee.SealedMessage) {},
		},
		{
			name:       "error - tampered ciphertext",
			openSecret: secret,
			mutate: func(msg *tee.SealedMessage) {
				msg.Ciphertext[len(msg.Ciphertext)-1] ^= 0xff
			},
			wantErr: tee.ErrHPKE,
		},
		{
			name:       "error - wrong secret",
			openSecret: bytes.Repeat([]byte{2}, tee.HPKEResponseSecretSize),
			mutate:     func(_ *tee.SealedMessage) {},
			wantErr:    tee.ErrHPKE,
		},
		{
			name:       "error - unexpected enc",
			openSecret: secret,
			mutate: func(msg *tee.SealedMessage) {
				msg.Enc = []byte("enc")
			},
			wantErr: tee.ErrHPKE,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			sealKey, err := tee.NewHPKEResponseKey(secret)
			require.NoError(t, err)
			openKey, err := tee.NewHPKEResponseKey(tt.openSecret)
			require.NoError(t, err)
			plaintext := []byte("reply")
			msg := sealKey.Seal(plaintext)
			tt.mutate(&msg)

			// when
			got, err := openKey.Open(msg)

			// then
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, plaintext, got)
		})
	}
}
