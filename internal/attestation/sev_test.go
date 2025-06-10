package attestation_test

import (
	_ "embed"
	"encoding/base64"
	"testing"
	"time"

	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/google/go-sev-guest/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/internal/attestation"
)

//go:embed testdata/sev-report-b64.txt
var sevReportB64 string

const (
	sevReportMeasurement          = "309f1b1e068fe6390234722a725c7a64f9d453cea13275a0d574f3bfc1f8864450e8c8f28a245fa1ed6ea6817ac85b2b"
	sevReportTimestampSeconds     = int64(1748808574)
	sevReportTimestampNanoseconds = int64(295000000)
)

func sevReportFromTestData(
	t *testing.T,
	reportB64 string,
	timestamp time.Time,
) ([]byte, *sevsnp.Report) {
	report, err := base64.StdEncoding.DecodeString(reportB64)
	require.NoError(t, err)

	pbReport, err := abi.ReportCertsToProto(report)
	require.NoError(t, err)

	opts := verify.DefaultOptions()
	opts.Now = timestamp
	err = verify.SnpAttestation(pbReport, opts)
	require.NoError(t, err)

	return report, pbReport.Report
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
		measurement := sevReportMeasurement
		timestamp := time.Unix(
			sevReportTimestampSeconds,
			sevReportTimestampNanoseconds,
		)
		report, _ := sevReportFromTestData(t, sevReportB64, timestamp)

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

	t.Run("error - invalid report", func(t *testing.T) {
		// given
		report := []byte("invalid attestation report")

		verifier, err := attestation.NewSEVVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report)

		// then
		assert.ErrorContains(t, err, "converting sev report to proto")
	})

	t.Run("error - expired report", func(t *testing.T) {
		// given
		timestamp := time.Unix(
			sevReportTimestampSeconds,
			sevReportTimestampNanoseconds,
		)
		report, _ := sevReportFromTestData(t, sevReportB64, timestamp)
		timestamp = time.Unix(0, 0)

		verifier, err := attestation.NewSEVVerifier()
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
			sevReportTimestampSeconds,
			sevReportTimestampNanoseconds,
		)
		report, _ := sevReportFromTestData(t, sevReportB64, timestamp)

		verifier, err := attestation.NewSEVVerifier()
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
		measurement := sevReportMeasurement
		timestamp := time.Unix(
			sevReportTimestampSeconds,
			sevReportTimestampNanoseconds,
		)
		report, _ := sevReportFromTestData(t, sevReportB64, timestamp)

		verifier, err := attestation.NewSEVVerifier()
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

func TestSEVIsDebugEnabled(t *testing.T) {
	t.Run("happy path - debug enabled", func(t *testing.T) {
		// given
		timestamp := time.Unix(
			sevReportTimestampSeconds,
			sevReportTimestampNanoseconds,
		)
		_, sevReport := sevReportFromTestData(t, sevReportB64, timestamp)

		// According to Pg. 31 of the SEV ABI Specification document, the 19th
		// bit of the Policy denotes whether debugging is enabled or not.
		// https://www.amd.com/content/dam/amd/en/documents/epyc-technical-docs/specifications/56860.pdf
		sevReport.Policy = sevReport.Policy | uint64(1<<19)

		// when
		got, err := attestation.SEVIsDebugEnabled(sevReport)

		// then
		assert.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("happy path - debug disabled", func(t *testing.T) {
		// given
		timestamp := time.Unix(
			sevReportTimestampSeconds,
			sevReportTimestampNanoseconds,
		)
		_, sevReport := sevReportFromTestData(t, sevReportB64, timestamp)

		// when
		got, err := attestation.SEVIsDebugEnabled(sevReport)

		// then
		assert.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("error - parsing policy", func(t *testing.T) {
		// given
		timestamp := time.Unix(
			sevReportTimestampSeconds,
			sevReportTimestampNanoseconds,
		)
		_, sevReport := sevReportFromTestData(t, sevReportB64, timestamp)
		sevReport.Policy = 0x0000000000000000

		// when
		_, err := attestation.SEVIsDebugEnabled(sevReport)

		// then
		assert.ErrorContains(t, err, "parsing policy")
	})
}

func TestSEVVerifyMeasurement(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		measurement := sevReportMeasurement
		timestamp := time.Unix(
			sevReportTimestampSeconds,
			sevReportTimestampNanoseconds,
		)
		_, sevReport := sevReportFromTestData(t, sevReportB64, timestamp)

		// when
		err := attestation.SEVVerifyMeasurement(measurement, sevReport)

		// then
		assert.NoError(t, err)
	})

	t.Run("happy path - no measurement", func(t *testing.T) {
		// given
		measurement := ""
		timestamp := time.Unix(
			sevReportTimestampSeconds,
			sevReportTimestampNanoseconds,
		)
		_, sevReport := sevReportFromTestData(t, sevReportB64, timestamp)

		// when
		err := attestation.SEVVerifyMeasurement(measurement, sevReport)

		// then
		assert.NoError(t, err)
	})

	t.Run("error - invalid measurement", func(t *testing.T) {
		// given
		measurement := "invalid measurement"
		timestamp := time.Unix(
			sevReportTimestampSeconds,
			sevReportTimestampNanoseconds,
		)
		_, sevReport := sevReportFromTestData(t, sevReportB64, timestamp)

		// when
		err := attestation.SEVVerifyMeasurement(measurement, sevReport)

		// then
		assert.ErrorContains(t, err, "parsing measurement")
	})

	t.Run("error - incorrect measurement", func(t *testing.T) {
		// given
		measurement := sevReportMeasurement
		measurement = measurement[:len(measurement)-2]
		timestamp := time.Unix(
			sevReportTimestampSeconds,
			sevReportTimestampNanoseconds,
		)
		_, sevReport := sevReportFromTestData(t, sevReportB64, timestamp)

		// when
		err := attestation.SEVVerifyMeasurement(measurement, sevReport)

		// then
		assert.ErrorContains(t, err, "mismatch: expected")
	})
}
