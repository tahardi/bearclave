package tee_test

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/tee"
)

func TestVerifyUserData(t *testing.T) {
	hash := sha256.Sum256([]byte("hello"))
	otherHash := sha256.Sum256([]byte("other"))
	nonZeroPadding := append(append([]byte{}, hash[:]...), make([]byte, 32)...)
	nonZeroPadding[len(nonZeroPadding)-1] = 0x01

	tests := []struct {
		name     string
		expected []byte
		userData []byte
		wantErr  error
	}{
		{
			name:     "happy path - exact 32 bytes",
			expected: hash[:],
			userData: []byte("hello"),
		},
		{
			name:     "happy path - zero padded to 64 bytes",
			expected: append(append([]byte{}, hash[:]...), make([]byte, 32)...),
			userData: []byte("hello"),
		},
		{
			name:     "error - short measurement",
			expected: hash[:16],
			userData: []byte("hello"),
			wantErr:  tee.ErrVerifier,
		},
		{
			name:     "error - non-zero padding",
			expected: nonZeroPadding,
			userData: []byte("hello"),
			wantErr:  tee.ErrVerifier,
		},
		{
			name:     "error - mismatch",
			expected: otherHash[:],
			userData: []byte("hello"),
			wantErr:  tee.ErrVerifier,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			err := tee.VerifyUserData(tt.expected, tt.userData)

			// then
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
