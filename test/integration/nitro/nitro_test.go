package nitro_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/internal/attestation"
	"github.com/tahardi/bearclave/internal/drivers"
)

const (
	nitroReportDebugMeasurementJSON = `{
  "pcrs": {
    "0": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
    "1": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
    "2": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
    "3": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
    "8": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
  },
}`
)

func TestNitro_Drivers(t *testing.T) {
	t.Run("happy path - new nsm client", func(t *testing.T) {
		client, err := drivers.NewNSMClient()
		require.NoError(t, err)
		require.NotNil(t, client)
		client.Close()
	})

	// NOTE: PCRs 0-15 are always locked regardless of whether you are running
	// in debug mode or not. In debug mode, every PCR is 0 except for PCR 4
	t.Run("happy path - describe pcr 0 (debug)", func(t *testing.T) {
		// given
		client, err := drivers.NewNSMClient()
		require.NoError(t, err)
		defer client.Close()

		index := uint16(0)

		// when
		value, lock, err := client.DescribePCR(index)

		// then
		require.NoError(t, err)
		require.NotEmpty(t, value)
		assert.True(t, lock)
		for i, v := range value {
			assert.Equal(t, byte(0), v, "val %d should be 0", i)
		}
	})

	t.Run("happy path - extend pcr", func(t *testing.T) {
		// given
		client, err := drivers.NewNSMClient()
		require.NoError(t, err)
		defer client.Close()

		index := uint16(16)
		data := []byte("Hello, World!")
		oldVal, lock, err := client.DescribePCR(index)
		require.NoError(t, err)
		require.NotEmpty(t, oldVal)
		require.False(t, lock)

		// when
		newVal, err := client.ExtendPCR(index, data)

		// then
		require.NoError(t, err)
		assert.NotEqual(t, oldVal, newVal)
	})

	t.Run("happy path - lock pcr", func(t *testing.T) {
		// given
		client, err := drivers.NewNSMClient()
		require.NoError(t, err)
		defer client.Close()

		index := uint16(16)
		_, lock, err := client.DescribePCR(index)
		require.NoError(t, err)
		require.False(t, lock)

		// when
		err = client.LockPCR(index)

		// then
		require.NoError(t, err)

		_, lock, err = client.DescribePCR(index)
		require.NoError(t, err)
		require.True(t, lock)
	})

	t.Run("happy path - lock pcrs", func(t *testing.T) {
		// given
		end := uint16(32)
		client, err := drivers.NewNSMClient()
		require.NoError(t, err)
		defer client.Close()

		// when
		err = client.LockPCRs(end)

		// then
		require.NoError(t, err)
		for i := uint16(0); i < end; i++ {
			_, lock, err := client.DescribePCR(i)
			require.NoError(t, err)
			require.True(t, lock, "PCR %d should be locked", i)
		}
	})

	t.Run("happy path - get attestation", func(t *testing.T) {
		// given
		client, err := drivers.NewNSMClient()
		require.NoError(t, err)
		defer client.Close()

		nonce := []byte("nonce")
		publickKey := []byte("public key")
		userData := []byte("Hello, world!")

		// when
		report, err := client.GetAttestation(nonce, publickKey, userData)

		// then
		require.NoError(t, err)
		assert.NotEmpty(t, report)
	})

	t.Run("happy path - get description", func(t *testing.T) {
		// given
		client, err := drivers.NewNSMClient()
		require.NoError(t, err)
		defer client.Close()

		// when
		description, err := client.GetDescription()

		// then
		require.NoError(t, err)
		assert.NotEmpty(t, description)
	})

	t.Run("happy path - get random", func(t *testing.T) {
		// given
		client, err := drivers.NewNSMClient()
		require.NoError(t, err)
		defer client.Close()

		wantLength := uint16(32)

		// when
		got, err := client.GetRandom(wantLength)

		// then
		require.NoError(t, err)
		assert.Len(t, got, int(wantLength))
	})
}

func TestNitro_Attestation(t *testing.T) {
	t.Run("happy path - attest & verify (debug)", func(t *testing.T) {
		// given
		nonce := []byte("nonce")
		userData := []byte("Hello, world!")
		attester, err := attestation.NewNitroAttester()
		require.NoError(t, err)
		defer attester.Close()

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		attested, err := attester.Attest(
			attestation.WithAttestNonce(nonce),
			attestation.WithAttestUserData(userData),
		)
		require.NoError(t, err)

		verified, err := verifier.Verify(
			attested,
			attestation.WithVerifyVerifyNonce(nonce),
			attestation.WithVerifyDebug(true),
		)

		// then
		require.NoError(t, err)
		assert.Equal(t, userData, verified.UserData)
	})
}
