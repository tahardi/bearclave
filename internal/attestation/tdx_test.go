package attestation_test

import (
	_ "embed"
	"encoding/base64"
	"testing"
	"time"

	"github.com/google/go-tdx-guest/abi"
	pb "github.com/google/go-tdx-guest/proto/tdx"
	"github.com/google/go-tdx-guest/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/internal/attestation"
)

//go:embed testdata/tdx-report-b64.txt
var tdxReportB64 string

const (
	tdxReportMeasurementJSON = `{
  "tee_tcb_svn": "CAEIAAAAAAAAAAAAAAAAAA==",
  "mr_seam": "v7NgrI5iM6G8oUM8r3OC2VwWW0p3+wC/FDXloI8wDN/q1e5oRhr9m2xyjc51NGAt",
  "mr_signer_seam": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
  "seam_attributes": "AAAAAAAAAAA=",
  "td_attributes": "AAAAEAAAAAA=",
  "xfam": "5wAGAAAAAAA=",
  "mr_td": "8nLYSS0x9v/6HQroHtLSQKLdS4Gl9evsfonJo195w9gxWI8Y068TqbM3OY75G7Nr",
  "mr_config_id": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
  "mr_owner": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
  "mr_owner_config": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
  "rtmrs": [
    "rRswApVNahPhBLWg69PzXOJ11lJZ6693qpzUBKv95YqdvGEgeoKPWoPbyCmD/ohO",
    "541YF6n1DTDLxiDYwkfmmWOptN82LDCu/6SnPGQ+c1dPtqSGHWuDqNwJWzkvN+Ae",
    "lfXvmOAaKoL3eRi2SxkYsp2ypIr4ng9XdCBy+HxNklgRzNdMwFQNavJaVnHC3/Nl",
    "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
  ]
}`
	tdxReportTimestampSeconds     = int64(1748808574)
	tdxReportTimestampNanoseconds = int64(295000000)
)

func tdxReportFromTestData(
	t *testing.T,
	reportB64 string,
	timestamp time.Time,
) ([]byte, *pb.QuoteV4) {
	report, err := base64.StdEncoding.DecodeString(reportB64)
	require.NoError(t, err)

	pbQuote, err := abi.QuoteToProto(report)
	require.NoError(t, err)

	opts := verify.DefaultOptions()
	opts.Now = timestamp
	err = verify.TdxQuote(pbQuote, opts)
	require.NoError(t, err)

	quoteV4, ok := pbQuote.(*pb.QuoteV4)
	require.True(t, ok)
	return report, quoteV4
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
		measurement := tdxReportMeasurementJSON
		timestamp := time.Unix(
			tdxReportTimestampSeconds,
			tdxReportTimestampNanoseconds,
		)
		report, _ := tdxReportFromTestData(t, tdxReportB64, timestamp)

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

	t.Run("error - invalid report", func(t *testing.T) {
		// given
		report := []byte("invalid attestation report")

		verifier, err := attestation.NewTDXVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report)

		// then
		assert.ErrorContains(t, err, "converting tdx report to proto")
	})

	t.Run("error - expired report", func(t *testing.T) {
		// given
		timestamp := time.Unix(
			tdxReportTimestampSeconds,
			tdxReportTimestampNanoseconds,
		)
		report, _ := tdxReportFromTestData(t, tdxReportB64, timestamp)
		timestamp = time.Unix(0, 0)

		verifier, err := attestation.NewTDXVerifier()
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
			tdxReportTimestampSeconds,
			tdxReportTimestampNanoseconds,
		)
		report, _ := tdxReportFromTestData(t, tdxReportB64, timestamp)

		verifier, err := attestation.NewTDXVerifier()
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
		measurement := tdxReportMeasurementJSON
		timestamp := time.Unix(
			tdxReportTimestampSeconds,
			tdxReportTimestampNanoseconds,
		)
		report, _ := tdxReportFromTestData(t, tdxReportB64, timestamp)

		verifier, err := attestation.NewTDXVerifier()
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

func TestTDXIsDebugEnabled(t *testing.T) {
	t.Run("happy path - debug enabled", func(t *testing.T) {
		// given
		timestamp := time.Unix(
			tdxReportTimestampSeconds,
			tdxReportTimestampNanoseconds,
		)
		_, quoteV4 := tdxReportFromTestData(t, tdxReportB64, timestamp)
		quoteV4.TdQuoteBody.TdAttributes[0] = 1

		// when
		got, err := attestation.TDXIsDebugEnabled(quoteV4)

		// then
		assert.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("happy path - debug disabled", func(t *testing.T) {
		// given
		timestamp := time.Unix(
			tdxReportTimestampSeconds,
			tdxReportTimestampNanoseconds,
		)
		_, quoteV4 := tdxReportFromTestData(t, tdxReportB64, timestamp)

		// when
		got, err := attestation.TDXIsDebugEnabled(quoteV4)

		// then
		assert.NoError(t, err)
		assert.False(t, got)
	})
}

func TestTDXVerifyMeasurement(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		measurement := tdxReportMeasurementJSON
		timestamp := time.Unix(
			tdxReportTimestampSeconds,
			tdxReportTimestampNanoseconds,
		)
		_, quoteV4 := tdxReportFromTestData(t, tdxReportB64, timestamp)

		// when
		err := attestation.TDXVerifyMeasurement(measurement, quoteV4.GetTdQuoteBody())

		// then
		assert.NoError(t, err)
	})

	t.Run("happy path - no measurement", func(t *testing.T) {
		// given
		measurement := ""
		timestamp := time.Unix(
			tdxReportTimestampSeconds,
			tdxReportTimestampNanoseconds,
		)
		_, quoteV4 := tdxReportFromTestData(t, tdxReportB64, timestamp)

		// when
		err := attestation.TDXVerifyMeasurement(measurement, quoteV4.GetTdQuoteBody())

		// then
		assert.NoError(t, err)
	})

	t.Run("error - invalid measurement", func(t *testing.T) {
		// given
		measurement := "invalid measurement"
		timestamp := time.Unix(
			tdxReportTimestampSeconds,
			tdxReportTimestampNanoseconds,
		)
		_, quoteV4 := tdxReportFromTestData(t, tdxReportB64, timestamp)

		// when
		err := attestation.TDXVerifyMeasurement(measurement, quoteV4.GetTdQuoteBody())

		// then
		assert.ErrorContains(t, err, "unmarshaling measurement")
	})

	t.Run("error - incorrect measurement", func(t *testing.T) {
		// given
		measurement := tdxReportMeasurementJSON
		timestamp := time.Unix(
			tdxReportTimestampSeconds,
			tdxReportTimestampNanoseconds,
		)
		_, quoteV4 := tdxReportFromTestData(t, tdxReportB64, timestamp)
		quoteV4.TdQuoteBody.TdAttributes[0] = 1

		// when
		err := attestation.TDXVerifyMeasurement(measurement, quoteV4.GetTdQuoteBody())

		// then
		assert.ErrorContains(t, err, "mismatch: expected")
	})
}
