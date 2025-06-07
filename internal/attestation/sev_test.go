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

//go:embed testdata/sev-attestation-b64.txt
var sevAttestationB64 string

const (
	sevAttestationMeasurement          = "309f1b1e068fe6390234722a725c7a64f9d453cea13275a0d574f3bfc1f8864450e8c8f28a245fa1ed6ea6817ac85b2b"
	sevAttestationTimestampSeconds     = int64(1748808574)
	sevAttestationTimestampNanoseconds = int64(295000000)
)

func sevAttestationFromTestData(t *testing.T) ([]byte, string, time.Time) {
	report, err := base64.StdEncoding.DecodeString(sevAttestationB64)
	require.NoError(t, err)

	timestamp := time.Unix(
		sevAttestationTimestampSeconds,
		sevAttestationTimestampNanoseconds,
	)

	return report, sevAttestationMeasurement, timestamp
}

func TestSEV_Interfaces(t *testing.T) {
	t.Run("Attester", func(t *testing.T) {
		var _ attestation.Attester = &attestation.SEVAttester{}
	})
	t.Run("Verifier", func(t *testing.T) {
		var _ attestation.Verifier = &attestation.SEVVerifier{}
	})
}

func TestSEVVerifier_Verify(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		want := []byte("Hello, world!")
		report, measurement, timestamp := sevAttestationFromTestData(t)

		verifier, err := attestation.NewSEVVerifier()
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
		report, _, timestamp := sevAttestationFromTestData(t)

		verifier, err := attestation.NewSEVVerifier()
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

		verifier, err := attestation.NewSEVVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report)

		// then
		assert.ErrorContains(t, err, "converting sev attestation to proto")
	})

	t.Run("error - expired report", func(t *testing.T) {
		// given
		timestamp := time.Unix(0, 0)
		report, _, _ := sevAttestationFromTestData(t)

		verifier, err := attestation.NewSEVVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report, attestation.WithTimestamp(timestamp))

		// then
		assert.ErrorContains(t, err, "certificate has expired or is not yet valid")
	})

	t.Run("error - invalid measurement format", func(t *testing.T) {
		// given
		invalidMeasurement := "invalid measurement format"
		report, _, timestamp := sevAttestationFromTestData(t)

		verifier, err := attestation.NewSEVVerifier()
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
		wrongMeasurement := "009f1b1e068fe6390234722a725c7a64f9d453cea13275a0d574f3bfc1f8864450e8c8f28a245fa1ed6ea6817ac85b2b"
		report, _, timestamp := sevAttestationFromTestData(t)

		verifier, err := attestation.NewSEVVerifier()
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
