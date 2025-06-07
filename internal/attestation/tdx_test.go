package attestation_test

import (
	_ "embed"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/internal/attestation"
)

//go:embed testdata/tdx-attestation-b64.txt
var tdxAttestationB64 string

const (
	tdxAttestationMrTD                 = "f272d8492d31f6fffa1d0ae81ed2d240a2dd4b81a5f5ebec7e89c9a35f79c3d831588f18d3af13a9b337398ef91bb36b"
	tdxAttestationTimestampSeconds     = int64(1748808574)
	tdxAttestationTimestampNanoseconds = int64(295000000)
)

func tdxAttestationFromTestData(t *testing.T) ([]byte, string, time.Time) {
	report, err := base64.StdEncoding.DecodeString(tdxAttestationB64)
	require.NoError(t, err)

	timestamp := time.Unix(
		tdxAttestationTimestampSeconds,
		tdxAttestationTimestampNanoseconds,
	)

	return report, tdxAttestationMrTD, timestamp
}

func TestTDX_Interfaces(t *testing.T) {
	t.Run("Attester", func(t *testing.T) {
		var _ attestation.Attester = &attestation.TDXAttester{}
	})
	t.Run("Verifier", func(t *testing.T) {
		var _ attestation.Verifier = &attestation.TDXVerifier{}
	})
}

func TestTDXVerifier_Verify(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		want := []byte("Hello, world!")
		report, measurement, timestamp := tdxAttestationFromTestData(t)

		verifier, err := attestation.NewTDXVerifier()
		require.NoError(t, err)

		// when
		got, err := verifier.Verify(
			report,
			attestation.WithMeasurement(measurement),
			attestation.WithTimestamp(timestamp),
		)

		// then
		assert.NoError(t, err)
		assert.Contains(t, string(got), string(want))
	})

	t.Run("happy path - no measurement", func(t *testing.T) {
		// given
		want := []byte("Hello, world!")
		report, _, timestamp := tdxAttestationFromTestData(t)

		verifier, err := attestation.NewTDXVerifier()
		require.NoError(t, err)

		// when
		got, err := verifier.Verify(report, attestation.WithTimestamp(timestamp))

		// then
		assert.NoError(t, err)
		assert.Contains(t, string(got), string(want))
	})

	t.Run("error - invalid attestation report", func(t *testing.T) {
		// given
		report := []byte("invalid attestation report")

		verifier, err := attestation.NewTDXVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report)

		// then
		assert.ErrorContains(t, err, "converting tdx attestation to proto")
	})

	t.Run("error - expired report", func(t *testing.T) {
		// given
		timestamp := time.Unix(0, 0)
		report, _, _ := tdxAttestationFromTestData(t)

		verifier, err := attestation.NewTDXVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report, attestation.WithTimestamp(timestamp))

		// then
		assert.ErrorContains(t, err, "certificate has expired or is not yet valid")
	})

	t.Run("error - invalid measurement format", func(t *testing.T) {
		// given
		invalidMeasurement := "invalid measurement format"
		report, _, timestamp := tdxAttestationFromTestData(t)

		verifier, err := attestation.NewTDXVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(
			report,
			attestation.WithMeasurement(invalidMeasurement),
			attestation.WithTimestamp(timestamp),
		)

		// then
		assert.ErrorContains(t, err, "decoding measurement")
	})

	t.Run("error - wrong measurement", func(t *testing.T) {
		// given
		wrongMeasurement := "0272d8492d31f6fffa1d0ae81ed2d240a2dd4b81a5f5ebec7e89c9a35f79c3d831588f18d3af13a9b337398ef91bb36b"
		report, _, timestamp := tdxAttestationFromTestData(t)

		verifier, err := attestation.NewTDXVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(
			report,
			attestation.WithMeasurement(wrongMeasurement),
			attestation.WithTimestamp(timestamp),
		)

		// then
		assert.ErrorContains(t, err, "measurement mismatch")
	})
}
