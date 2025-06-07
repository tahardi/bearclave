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
  "0": "1606040ac5afb19824cb0f783ed0f90583ef1ac555dc3b0b5f00d564ec0d206bd7a56ca6ee049ba2e03bb7d4ea88d00b",
  "1": "4b4d5b3661b3efc12920900c80e126e4ce783c522de6c02a2a5bf7af3a2b9327b86776f188e4be1c1c404a129dbda493",
  "2": "e0742e0e40b8576aeadccff8539e448399b9924bb9701772bb5d2501d97757faba6d2aa390e761ef60c5a4d7452f101f",
  "3": "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "4": "a823da6c81d753e9e119c65d34e961eadeadb3ce6f3e95db1214716357fe6b32dc02f6a16e0b0137eb0a6c27e713ecaa",
  "8": "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
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
  "0": "1606040ac5afb19824cb0f783ed0f90583ef1ac555dc3b0b5f00d564ec0d206bd7a56ca6ee049ba2e03bb7d4ea88d00b",
  "1": "4b4d5b3661b3efc12920900c80e126e4ce783c522de6c02a2a5bf7af3a2b9327b86776f188e4be1c1c404a129dbda493",
  "2": "e0742e0e40b8576aeadccff8539e448399b9924bb9701772bb5d2501d97757faba6d2aa390e761ef60c5a4d7452f101f",
  "3": "100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "4": "a823da6c81d753e9e119c65d34e961eadeadb3ce6f3e95db1214716357fe6b32dc02f6a16e0b0137eb0a6c27e713ecaa",
  "8": "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
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
