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
	sevReportMeasurementJSON = `{
  "version": 4,
  "guest_svn": 0,
  "policy": 196608,
  "family_id": "AAAAAAAAAAAAAAAAAAAAAA==",
  "image_id": "AAAAAAAAAAAAAAAAAAAAAA==",
  "vmpl": 0,
  "current_tcb": 15787649968723984388,
  "platform_info": 37,
  "signer_info": 0,
  "measurement": "MJ8bHgaP5jkCNHIqclx6ZPnUU86hMnWg1XTzv8H4hkRQ6MjyiiRfoe1upoF6yFsr",
  "host_data": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "id_key_digest": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
  "author_key_digest": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
  "report_id": "6FRblnt1xgkavHxgGL8mWw8B6gEO+YklgWmboVVrI9M=",
  "report_id_ma": "//////////////////////////////////////////8=",
  "reported_tcb": 15787649968723984388,
  "chip_id": "Etk3ouD6ZGkppCuDr+W0eNOYdzXjhKAVcH7AaJjhfAmXu4cuC7tyz9ccUMoVYHZ84eGYcz5kjA7APC1nqsjAuQ==",
  "committed_tcb": 15787649968723984388,
  "current_build": 31,
  "current_minor": 55,
  "current_major": 1,
  "committed_build": 31,
  "committed_minor": 55,
  "committed_major": 1,
  "launch_tcb": 15787649968723984388,
  "cpuid_1eax_fms": 10489617
}`
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
		measurement := sevReportMeasurementJSON
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
		assert.ErrorContains(t, err, "verifying measurement")
	})

	t.Run("error - debug mode mismatch", func(t *testing.T) {
		// given
		measurement := sevReportMeasurementJSON
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
		measurement := sevReportMeasurementJSON
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
		assert.ErrorContains(t, err, "unmarshaling measurement")
	})

	t.Run("error - incorrect measurement", func(t *testing.T) {
		// given
		measurement := sevReportMeasurementJSON
		timestamp := time.Unix(
			sevReportTimestampSeconds,
			sevReportTimestampNanoseconds,
		)
		_, sevReport := sevReportFromTestData(t, sevReportB64, timestamp)
		sevReport.Measurement[0] = 0

		// when
		err := attestation.SEVVerifyMeasurement(measurement, sevReport)

		// then
		assert.ErrorContains(t, err, "measurement mismatch: expected")
	})
}
