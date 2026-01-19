package tdx_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/internal/attestation"
	"github.com/tahardi/bearclave/internal/drivers"
)

func TestTDX_Drivers(t *testing.T) {
	t.Run("happy path - new tdx client", func(t *testing.T) {
		client, err := drivers.NewTDXClient()
		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("happy path - get report", func(t *testing.T) {
		// given
		client, err := drivers.NewTDXClient()
		require.NoError(t, err)

		userData := []byte("Hello, World!")

		// when
		report, err := client.GetReport(userData)

		// then
		require.NoError(t, err)
		require.NotEmpty(t, report)
	})
}

func TestTDX_Attestation(t *testing.T) {
	t.Run("happy path - attest & verify", func(t *testing.T) {
		// given
		userData := []byte("Hello, World!")
		attester, err := attestation.NewTDXAttester()
		require.NoError(t, err)

		verifier, err := attestation.NewTDXVerifier()
		require.NoError(t, err)

		// when
		attested, err := attester.Attest(attestation.WithAttestUserData(userData))
		require.NoError(t, err)

		verified, err := verifier.Verify(attested)

		// then
		require.NoError(t, err)
		assert.True(t, bytes.Contains(verified.UserData, userData))
	})
}
