package attestation_test

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/hf/nitrite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/internal/attestation"
)

//go:embed testdata/nitro-report-b64.txt
var nitroReportB64 string

//go:embed testdata/nitro-report-debug-b64.txt
var nitroReportDebugB64 string

const (
	nitroReportMeasurementJSON = `{
  "pcrs": {
    "0": "FgYECsWvsZgkyw94PtD5BYPvGsVV3DsLXwDVZOwNIGvXpWym7gSbouA7t9TqiNAL",
    "1": "S01bNmGz78EpIJAMgOEm5M54PFIt5sAqKlv3rzorkye4Z3bxiOS+HBxAShKdvaST",
    "2": "4HQuDkC4V2rq3M/4U55Eg5m5kku5cBdyu10lAdl3V/q6bSqjkOdh72DFpNdFLxAf",
    "3": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
    "4": "qCPabIHXU+nhGcZdNOlh6t6ts85vPpXbEhRxY1f+azLcAvahbgsBN+sKbCfnE+yq",
    "8": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
  },
  "module_id": "i-01bdf23ce28366cb5-enc01974a1e041bde39"
}`
	nitroReportDebugMeasurementJSON = `{
  "pcrs": {
    "0": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
    "1": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
    "2": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
    "3": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
    "4": "qCPabIHXU+nhGcZdNOlh6t6ts85vPpXbEhRxY1f+azLcAvahbgsBN+sKbCfnE+yq",
    "8": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
  },
  "module_id": "i-01bdf23ce28366cb5-enc019759c92189a1e4"
}`
	nitroReportTimestampSeconds          = int64(1749295504)
	nitroReportTimestampNanoseconds      = int64(541000000)
	nitroReportDebugTimestampSeconds     = int64(1749558205)
	nitroReportDebugTimestampNanoseconds = int64(687000000)
)

func nitroReportFromTestData(
	t *testing.T,
	reportB64 string,
	timestamp time.Time,
) (*attestation.AttestResult, *nitrite.Document) {
	t.Helper()
	reportBytes, err := base64.StdEncoding.DecodeString(reportB64)
	require.NoError(t, err)

	opts := nitrite.VerifyOptions{CurrentTime: timestamp}
	result, err := nitrite.Verify(reportBytes, opts)
	require.NoError(t, err)
	return &attestation.AttestResult{Report: reportBytes}, result.Document
}

func TestNitro_Interfaces(t *testing.T) {
	t.Run("Attester", func(_ *testing.T) {
		var _ attestation.Attester = &attestation.NitroAttester{}
	})
	t.Run("Verifier", func(_ *testing.T) {
		var _ attestation.Verifier = &attestation.NitroVerifier{}
	})
}

func TestNitroVerifier_Verify(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		want := []byte("Hello, world!")
		measurement := nitroReportMeasurementJSON
		timestamp := time.Unix(
			nitroReportTimestampSeconds,
			nitroReportTimestampNanoseconds,
		)
		report, _ := nitroReportFromTestData(t, nitroReportB64, timestamp)

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		got, err := verifier.Verify(
			report,
			attestation.WithVerifyMeasurement(measurement),
			attestation.WithVerifyTimestamp(timestamp),
		)

		// then
		require.NoError(t, err)
		assert.Equal(t, want, got.UserData)
	})

	t.Run("happy path - debug", func(t *testing.T) {
		// given
		want := []byte("Hello, world!")
		measurement := nitroReportDebugMeasurementJSON
		timestamp := time.Unix(
			nitroReportDebugTimestampSeconds,
			nitroReportDebugTimestampNanoseconds,
		)
		report, _ := nitroReportFromTestData(t, nitroReportDebugB64, timestamp)

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		got, err := verifier.Verify(
			report,
			attestation.WithVerifyDebug(true),
			attestation.WithVerifyMeasurement(measurement),
			attestation.WithVerifyTimestamp(timestamp),
		)

		// then
		require.NoError(t, err)
		assert.Equal(t, want, got.UserData)
	})

	t.Run("error - invalid report", func(t *testing.T) {
		// given
		report := &attestation.AttestResult{Report: []byte("invalid attestation report")}

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifier)
		assert.ErrorContains(t, err, "verifying report")
	})

	t.Run("error - expired report", func(t *testing.T) {
		// given
		timestamp := time.Unix(
			nitroReportTimestampSeconds,
			nitroReportTimestampNanoseconds,
		)
		report, _ := nitroReportFromTestData(t, nitroReportB64, timestamp)
		timestamp = time.Unix(0, 0)

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report, attestation.WithVerifyTimestamp(timestamp))

		// then
		require.ErrorIs(t, err, attestation.ErrVerifier)
		assert.ErrorContains(t, err, "certificate has expired or is not yet valid")
	})

	t.Run("error - verifying measurement", func(t *testing.T) {
		// given
		measurement := "invalid measurement"
		timestamp := time.Unix(
			nitroReportTimestampSeconds,
			nitroReportTimestampNanoseconds,
		)
		report, _ := nitroReportFromTestData(t, nitroReportB64, timestamp)

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(
			report,
			attestation.WithVerifyMeasurement(measurement),
			attestation.WithVerifyTimestamp(timestamp),
		)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifierMeasurement)
	})

	t.Run("error - debug mode mismatch", func(t *testing.T) {
		// given
		measurement := nitroReportMeasurementJSON
		timestamp := time.Unix(
			nitroReportTimestampSeconds,
			nitroReportTimestampNanoseconds,
		)
		report, _ := nitroReportFromTestData(t, nitroReportB64, timestamp)

		verifier, err := attestation.NewNitroVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(
			report,
			attestation.WithVerifyDebug(true),
			attestation.WithVerifyMeasurement(measurement),
			attestation.WithVerifyTimestamp(timestamp),
		)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifierDebugMode)
	})
}

func TestNitroIsDebugEnabled(t *testing.T) {
	t.Run("happy path - debug enabled", func(t *testing.T) {
		// given
		document := &nitrite.Document{}
		document.PCRs = make(map[uint][]byte, 3)
		for i := range 3 {
			document.PCRs[uint(i)] = make([]byte, 32)
		}

		// when
		got, err := attestation.NitroIsDebugEnabled(document)

		// then
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("happy path - debug disabled", func(t *testing.T) {
		// given
		document := &nitrite.Document{}
		document.PCRs = make(map[uint][]byte, 3)
		for i := range 3 {
			document.PCRs[uint(i)] = make([]byte, 32)
			document.PCRs[uint(i)][0] = 1
		}

		// when
		got, err := attestation.NitroIsDebugEnabled(document)

		// then
		require.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("error - no pcrs provided", func(t *testing.T) {
		// given
		document := &nitrite.Document{}
		document.PCRs = make(map[uint][]byte, 0)

		// when
		_, err := attestation.NitroIsDebugEnabled(document)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifierDebugMode)
		assert.ErrorContains(t, err, "no pcrs provided")
	})

	t.Run("error - missing pcr", func(t *testing.T) {
		// given
		document := &nitrite.Document{}
		document.PCRs = make(map[uint][]byte, 2)
		for i := range 2 {
			document.PCRs[uint(i)] = make([]byte, 32)
		}

		// when
		_, err := attestation.NitroIsDebugEnabled(document)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifierDebugMode)
		assert.ErrorContains(t, err, "missing pcr '2'")
	})
}

func TestNitroVerifyMeasurement(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		measurementJSON := nitroReportMeasurementJSON
		timestamp := time.Unix(
			nitroReportTimestampSeconds,
			nitroReportTimestampNanoseconds,
		)
		_, document := nitroReportFromTestData(t, nitroReportB64, timestamp)

		// when
		err := attestation.NitroVerifyMeasurement(measurementJSON, document)

		// then
		assert.NoError(t, err)
	})

	t.Run("happy path - no measurement", func(t *testing.T) {
		// given
		measurementJSON := ""
		timestamp := time.Unix(
			nitroReportTimestampSeconds,
			nitroReportTimestampNanoseconds,
		)
		_, document := nitroReportFromTestData(t, nitroReportB64, timestamp)

		// when
		err := attestation.NitroVerifyMeasurement(measurementJSON, document)

		// then
		assert.NoError(t, err)
	})

	t.Run("happy path - no module ID", func(t *testing.T) {
		// given
		measurement := attestation.NitroMeasurement{}
		measurementJSON, err := json.Marshal(measurement)
		require.NoError(t, err)

		timestamp := time.Unix(
			nitroReportTimestampSeconds,
			nitroReportTimestampNanoseconds,
		)
		_, document := nitroReportFromTestData(t, nitroReportB64, timestamp)

		// when
		err = attestation.NitroVerifyMeasurement(string(measurementJSON), document)

		// then
		assert.NoError(t, err)
	})

	t.Run("error - invalid measurement format", func(t *testing.T) {
		// given
		measurementJSON := "invalid measurement format"
		timestamp := time.Unix(
			nitroReportTimestampSeconds,
			nitroReportTimestampNanoseconds,
		)
		_, document := nitroReportFromTestData(t, nitroReportB64, timestamp)

		// when
		err := attestation.NitroVerifyMeasurement(measurementJSON, document)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifierMeasurement)
		assert.ErrorContains(t, err, "unmarshaling measurement")
	})

	t.Run("error - missing pcr", func(t *testing.T) {
		// given
		measurementJSON := nitroReportMeasurementJSON
		timestamp := time.Unix(
			nitroReportTimestampSeconds,
			nitroReportTimestampNanoseconds,
		)
		_, document := nitroReportFromTestData(t, nitroReportB64, timestamp)
		document.PCRs = make(map[uint][]byte, 0)

		// when
		err := attestation.NitroVerifyMeasurement(measurementJSON, document)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifierMeasurement)
		assert.ErrorContains(t, err, "missing pcr")
	})

	t.Run("error - incorrect measurement", func(t *testing.T) {
		// given
		measurementJSON := nitroReportMeasurementJSON
		timestamp := time.Unix(
			nitroReportTimestampSeconds,
			nitroReportTimestampNanoseconds,
		)
		_, document := nitroReportFromTestData(t, nitroReportB64, timestamp)
		document.PCRs[0][0] = 0

		// when
		err := attestation.NitroVerifyMeasurement(measurementJSON, document)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifierMeasurement)
		assert.ErrorContains(t, err, "mismatch: expected")
	})

	t.Run("error - incorrect module ID", func(t *testing.T) {
		// given
		measurementJSON := nitroReportMeasurementJSON
		timestamp := time.Unix(
			nitroReportTimestampSeconds,
			nitroReportTimestampNanoseconds,
		)
		_, document := nitroReportFromTestData(t, nitroReportB64, timestamp)
		document.ModuleID = "invalid module ID"

		// when
		err := attestation.NitroVerifyMeasurement(measurementJSON, document)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifierMeasurement)
		assert.ErrorContains(t, err, "mismatch: expected")
	})
}

func TestNitroVerifyNonce(t *testing.T) {
	t.Run("happy path - without nonce", func(t *testing.T) {
		// given
		document := &nitrite.Document{}

		// when
		err := attestation.NitroVerifyNonce(nil, document)

		// then
		assert.NoError(t, err)
	})

	t.Run("happy path - with nonce", func(t *testing.T) {
		// given
		nonce := []byte("nonce")
		document := &nitrite.Document{Nonce: nonce}

		// when
		err := attestation.NitroVerifyNonce(nonce, document)

		// then
		assert.NoError(t, err)
	})

	t.Run("error - nonce mismatch", func(t *testing.T) {
		// given
		nonce := []byte("nonce")
		document := &nitrite.Document{Nonce: nonce}

		// when
		err := attestation.NitroVerifyNonce([]byte("wrong nonce"), document)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifierNonce)
		assert.ErrorContains(t, err, "nonce mismatch")
	})
}
