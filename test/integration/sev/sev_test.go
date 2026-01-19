package sev_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/internal/attestation"
	"github.com/tahardi/bearclave/internal/drivers"
)

func TestSEV_Drivers(t *testing.T) {
	t.Run("happy path - new sev client", func(t *testing.T) {
		client, err := drivers.NewSEVClient()
		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("happy path - get report", func(t *testing.T) {
		// given
		client, err := drivers.NewSEVClient()
		require.NoError(t, err)

		userData := []byte("Hello, World!")

		// when
		report, err := client.GetReport(
			drivers.WithSEVReportUserData(userData),
			drivers.WithSEVReportCertTable(true),
		)

		// then
		require.NoError(t, err)
		require.NotEmpty(t, report)
	})
}

func TestSEV_Attestation(t *testing.T) {
	t.Run("happy path - attest & verify", func(t *testing.T) {
		// given
		userData := []byte("Hello, World!")
		attester, err := attestation.NewSEVAttester()
		require.NoError(t, err)

		verifier, err := attestation.NewSEVVerifier()
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