package unsafe_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/internal/unsafe"
)

func TestVerifier_Verify(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		want := []byte("hello world")
		attester, err := unsafe.NewAttester()
		require.NoError(t, err)

		attestation, err := attester.Attest(want)
		require.NoError(t, err)

		verifier, err := unsafe.NewVerifier()
		require.NoError(t, err)

		// when
		got, err := verifier.Verify(attestation)

		// then
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("error - invalid attestation", func(t *testing.T) {
		// given
		attestation := []byte("invalid attestation")

		verifier, err := unsafe.NewVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(attestation)

		// then
		assert.ErrorContains(t, err, "unmarshaling attestation")
	})

	t.Run("error - invalid signature", func(t *testing.T) {
		// given
		attestation := unsafe.Attestation{
			Userdata:  []byte("hello world"),
			Publickey: []byte("public key"),
			Signature: []byte("invalid signature"),
		}
		attestationBytes, err := json.Marshal(attestation)
		require.NoError(t, err)

		verifier, err := unsafe.NewVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(attestationBytes)

		// then
		assert.ErrorContains(t, err, "invalid signature")
	})
}
