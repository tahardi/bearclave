package notee_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/internal/attestation/notee"
)

func TestVerifier_Verify(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		want := []byte("hello world")
		attester, err := notee.NewAttester()
		require.NoError(t, err)

		attestation, err := attester.Attest(want)
		require.NoError(t, err)

		verifier, err := notee.NewVerifier()
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

		verifier, err := notee.NewVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(attestation)

		// then
		assert.ErrorContains(t, err, "unmarshaling attestation")
	})

	t.Run("error - invalid signature", func(t *testing.T) {
		// given
		attestation := notee.Attestation{
			Userdata:  []byte("hello world"),
			Publickey: []byte("public key"),
			Signature: []byte("invalid signature"),
		}
		attestationBytes, err := json.Marshal(attestation)
		require.NoError(t, err)

		verifier, err := notee.NewVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(attestationBytes)

		// then
		assert.ErrorContains(t, err, "invalid signature")
	})
}
