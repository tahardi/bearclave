package attestation_test

import (
	_ "embed"
	"encoding/base64"
	"testing"
	"time"

	"github.com/hf/nitrite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/internal/attestation"
)

//go:embed testdata/nitro-attestation-b64.txt
var nitroAttestationB64 string

//go:embed testdata/nitro-attestation-debug-b64.txt
var nitroAttestationDebugB64 string

const (
	nitroAttestationPCRsJSON = `{
  "0": "1606040ac5afb19824cb0f783ed0f90583ef1ac555dc3b0b5f00d564ec0d206bd7a56ca6ee049ba2e03bb7d4ea88d00b",
  "1": "4b4d5b3661b3efc12920900c80e126e4ce783c522de6c02a2a5bf7af3a2b9327b86776f188e4be1c1c404a129dbda493",
  "2": "e0742e0e40b8576aeadccff8539e448399b9924bb9701772bb5d2501d97757faba6d2aa390e761ef60c5a4d7452f101f",
  "3": "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "4": "a823da6c81d753e9e119c65d34e961eadeadb3ce6f3e95db1214716357fe6b32dc02f6a16e0b0137eb0a6c27e713ecaa",
  "8": "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
}`
	nitroAttestationDebugPCRsJSON = `{
  "0": "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "1": "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "2": "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "3": "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  "4": "a823da6c81d753e9e119c65d34e961eadeadb3ce6f3e95db1214716357fe6b32dc02f6a16e0b0137eb0a6c27e713ecaa",
  "8": "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
}`
	nitroAttestationTimestampSeconds          = int64(1749295504)
	nitroAttestationTimestampNanoseconds      = int64(541000000)
	nitroAttestationDebugTimestampSeconds     = int64(1749558205)
	nitroAttestationDebugTimestampNanoseconds = int64(687000000)
)

func nitroReportFromTestData(
	t *testing.T,
	reportB64 string,
	timestamp time.Time,
) ([]byte, *nitrite.Document) {
	reportBytes, err := base64.StdEncoding.DecodeString(reportB64)
	require.NoError(t, err)

	opts := nitrite.VerifyOptions{CurrentTime: timestamp}
	result, err := nitrite.Verify(reportBytes, opts)
	require.NoError(t, err)
	return reportBytes, result.Document
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
		measurement := nitroAttestationPCRsJSON
		timestamp := time.Unix(
			nitroAttestationTimestampSeconds,
			nitroAttestationTimestampNanoseconds,
		)
		report, _ := nitroReportFromTestData(t, nitroAttestationB64, timestamp)

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

	t.Run("happy path - debug", func(t *testing.T) {
		// given
		want := []byte("Hello, world!")
		measurement := nitroAttestationDebugPCRsJSON
		timestamp := time.Unix(
			nitroAttestationDebugTimestampSeconds,
			nitroAttestationDebugTimestampNanoseconds,
		)
		report, _ := nitroReportFromTestData(t, nitroAttestationDebugB64, timestamp)

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		got, err := verifier.Verify(
			report,
			attestation.WithDebug(true),
			attestation.WithMeasurement(measurement),
			attestation.WithTimestamp(timestamp),
		)

		// then
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("error - invalid report", func(t *testing.T) {
		// given
		report := []byte("invalid attestation report")

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report)

		// then
		assert.ErrorContains(t, err, "verifying report")
	})

	t.Run("error - expired report", func(t *testing.T) {
		// given
		timestamp := time.Unix(
			nitroAttestationTimestampSeconds,
			nitroAttestationTimestampNanoseconds,
		)
		report, _ := nitroReportFromTestData(t, nitroAttestationB64, timestamp)
		timestamp = time.Unix(0, 0)

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report, attestation.WithTimestamp(timestamp))

		// then
		assert.ErrorContains(t, err, "certificate has expired or is not yet valid")
	})

	t.Run("error - verifying measurement", func(t *testing.T) {
		// given
		measurement := "invalid measurement"
		timestamp := time.Unix(
			nitroAttestationTimestampSeconds,
			nitroAttestationTimestampNanoseconds,
		)
		report, _ := nitroReportFromTestData(t, nitroAttestationB64, timestamp)

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(
			report,
			attestation.WithMeasurement(measurement),
			attestation.WithTimestamp(timestamp),
		)

		// then
		assert.ErrorContains(t, err, "parsing measurement")
	})

	t.Run("error - debug mode mismatch", func(t *testing.T) {
		// given
		measurement := nitroAttestationPCRsJSON
		timestamp := time.Unix(
			nitroAttestationTimestampSeconds,
			nitroAttestationTimestampNanoseconds,
		)
		report, _ := nitroReportFromTestData(t, nitroAttestationB64, timestamp)

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(
			report,
			attestation.WithDebug(true),
			attestation.WithMeasurement(measurement),
			attestation.WithTimestamp(timestamp),
		)

		// then
		assert.ErrorContains(t, err, "debug mode mismatch")
	})
}

func TestNitroIsDebugEnabled(t *testing.T) {
	t.Run("happy path - debug enabled", func(t *testing.T) {
		// given
		document := &nitrite.Document{}
		document.PCRs = make(map[uint][]byte, 3)
		for i := 0; i < 3; i++ {
			document.PCRs[uint(i)] = make([]byte, 32)
		}

		// when
		got, err := attestation.NitroIsDebugEnabled(document)

		// then
		assert.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("happy path - debug disabled", func(t *testing.T) {
		// given
		document := &nitrite.Document{}
		document.PCRs = make(map[uint][]byte, 3)
		for i := 0; i < 3; i++ {
			document.PCRs[uint(i)] = make([]byte, 32)
			document.PCRs[uint(i)][0] = 1
		}

		// when
		got, err := attestation.NitroIsDebugEnabled(document)

		// then
		assert.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("error - no pcrs provided", func(t *testing.T) {
		// given
		document := &nitrite.Document{}
		document.PCRs = make(map[uint][]byte, 0)

		// when
		_, err := attestation.NitroIsDebugEnabled(document)

		// then
		assert.ErrorContains(t, err, "no pcrs provided")
	})

	t.Run("error - missing pcr", func(t *testing.T) {
		// given
		document := &nitrite.Document{}
		document.PCRs = make(map[uint][]byte, 2)
		for i := 0; i < 2; i++ {
			document.PCRs[uint(i)] = make([]byte, 32)
		}

		// when
		_, err := attestation.NitroIsDebugEnabled(document)

		// then
		assert.ErrorContains(t, err, "missing pcr '2'")
	})
}

func TestNitroVerifyMeasurement(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		measurement := nitroAttestationPCRsJSON
		timestamp := time.Unix(
			nitroAttestationTimestampSeconds,
			nitroAttestationTimestampNanoseconds,
		)
		_, document := nitroReportFromTestData(t, nitroAttestationB64, timestamp)

		// when
		err := attestation.NitroVerifyMeasurement(measurement, document)

		// then
		assert.NoError(t, err)
	})

	t.Run("happy path - no measurement", func(t *testing.T) {
		// given
		measurement := ""
		timestamp := time.Unix(
			nitroAttestationTimestampSeconds,
			nitroAttestationTimestampNanoseconds,
		)
		_, document := nitroReportFromTestData(t, nitroAttestationB64, timestamp)

		// when
		err := attestation.NitroVerifyMeasurement(measurement, document)

		// then
		assert.NoError(t, err)
	})

	t.Run("error - invalid measurement format", func(t *testing.T) {
		// given
		measurement := "invalid measurement format"
		timestamp := time.Unix(
			nitroAttestationTimestampSeconds,
			nitroAttestationTimestampNanoseconds,
		)
		_, document := nitroReportFromTestData(t, nitroAttestationB64, timestamp)

		// when
		err := attestation.NitroVerifyMeasurement(measurement, document)

		// then
		assert.ErrorContains(t, err, "parsing measurement")
	})

	t.Run("error - missing pcr", func(t *testing.T) {
		// given
		measurement := nitroAttestationPCRsJSON
		timestamp := time.Unix(
			nitroAttestationTimestampSeconds,
			nitroAttestationTimestampNanoseconds,
		)
		_, document := nitroReportFromTestData(t, nitroAttestationB64, timestamp)
		document.PCRs = make(map[uint][]byte, 0)

		// when
		err := attestation.NitroVerifyMeasurement(measurement, document)

		// then
		assert.ErrorContains(t, err, "missing pcr")
	})

	t.Run("error - incorrect measurement", func(t *testing.T) {
		// given
		measurement := nitroAttestationPCRsJSON
		timestamp := time.Unix(
			nitroAttestationTimestampSeconds,
			nitroAttestationTimestampNanoseconds,
		)
		_, document := nitroReportFromTestData(t, nitroAttestationB64, timestamp)
		document.PCRs[0][0] = 0

		// when
		err := attestation.NitroVerifyMeasurement(measurement, document)

		// then
		assert.ErrorContains(t, err, "mismatch: expected")
	})
}
