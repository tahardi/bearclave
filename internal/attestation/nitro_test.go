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

//go:embed testdata/nitro-attestation-b64.txt
var nitroAttestationB64 string

const (
	nitroAttestationPCRsJSON = `{
  "0": [22, 6, 4, 10, 197, 175, 177, 152, 36, 203, 15, 120, 62, 208, 249, 5, 131, 239, 26, 197, 85, 220, 59, 11, 95, 0, 213, 100, 236, 13, 32, 107, 215, 165, 108, 166, 238, 4, 155, 162, 224, 59, 183, 212, 234, 136, 208, 11],
  "1": [75, 77, 91, 54, 97, 179, 239, 193, 41, 32, 144, 12, 128, 225, 38, 228, 206, 120, 60, 82, 45, 230, 192, 42, 42, 91, 247, 175, 58, 43, 147, 39, 184, 103, 118, 241, 136, 228, 190, 28, 28, 64, 74, 18, 157, 189, 164, 147],
  "2": [224, 116, 46, 14, 64, 184, 87, 106, 234, 220, 207, 248, 83, 158, 68, 131, 153, 185, 146, 75, 185, 112, 23, 114, 187, 93, 37, 1, 217, 119, 87, 250, 186, 109, 42, 163, 144, 231, 97, 239, 96, 197, 164, 215, 69, 47, 16, 31],
  "3": [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
  "4": [168, 35, 218, 108, 129, 215, 83, 233, 225, 25, 198, 93, 52, 233, 97, 234, 222, 173, 179, 206, 111, 62, 149, 219, 18, 20, 113, 99, 87, 254, 107, 50, 220, 2, 246, 161, 110, 11, 1, 55, 235, 10, 108, 39, 231, 19, 236, 170],
  "8": [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]
}`
	nitroAttestationTimestampSeconds     = int64(1749295504)
	nitroAttestationTimestampNanoseconds = int64(541000000)
)

func nitroAttestationFromTestData(t *testing.T) ([]byte, string, time.Time) {
	report, err := base64.StdEncoding.DecodeString(nitroAttestationB64)
	require.NoError(t, err)

	timestamp := time.Unix(
		nitroAttestationTimestampSeconds,
		nitroAttestationTimestampNanoseconds,
	)

	return report, nitroAttestationPCRsJSON, timestamp
}

func TestNitro_Interfaces(t *testing.T) {
	t.Run("Attester", func(t *testing.T) {
		var _ attestation.Attester = &attestation.NitroAttester{}
	})
	t.Run("Verifier", func(t *testing.T) {
		var _ attestation.Verifier = &attestation.NitroVerifier{}
	})
}

func TestNitroVerifier_Verify(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		want := []byte("Hello, world!")
		report, measurement, timestamp := nitroAttestationFromTestData(t)

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		got, err := verifier.Verify(
			report,
			attestation.WithMeasurement(measurement),
			attestation.WithTimestamp(timestamp),
		)

		// then
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("happy path - no measurement", func(t *testing.T) {
		// given
		want := []byte("Hello, world!")
		report, _, timestamp := nitroAttestationFromTestData(t)

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		got, err := verifier.Verify(report, attestation.WithTimestamp(timestamp))

		// then
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("error - invalid attestation report", func(t *testing.T) {
		// given
		report := []byte("invalid attestation report")

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report)

		// then
		assert.ErrorContains(t, err, "verifying attestation")
	})

	t.Run("error - expired report", func(t *testing.T) {
		// given
		timestamp := time.Unix(0, 0)
		report, _, _ := nitroAttestationFromTestData(t)

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report, attestation.WithTimestamp(timestamp))

		// then
		assert.ErrorContains(t, err, "certificate has expired or is not yet valid")
	})

	t.Run("error - invalid measurement format", func(t *testing.T) {
		// given
		invalidMeasurement := "invalid measurement format"
		report, _, timestamp := nitroAttestationFromTestData(t)

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(
			report,
			attestation.WithMeasurement(invalidMeasurement),
			attestation.WithTimestamp(timestamp),
		)

		// then
		assert.ErrorContains(t, err, "unmarshaling pcrs")
	})

	t.Run("error - wrong measurement", func(t *testing.T) {
		// given
		wrongMeasurement := `{
  "0": [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
  "1": [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
  "2": [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
  "3": [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
  "4": [0, 35, 218, 108, 129, 215, 83, 233, 225, 25, 198, 93, 52, 233, 97, 234, 222, 173, 179, 206, 111, 62, 149, 219, 18, 20, 113, 99, 87, 254, 107, 50, 220, 2, 246, 161, 110, 11, 1, 55, 235, 10, 108, 39, 231, 19, 236, 170],
  "8": [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]
}`
		report, _, timestamp := nitroAttestationFromTestData(t)

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(
			report,
			attestation.WithMeasurement(wrongMeasurement),
			attestation.WithTimestamp(timestamp),
		)

		// then
		assert.ErrorContains(t, err, "verifying pcrs")
	})
}
