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
) (*attestation.AttestResult, *pb.QuoteV4) {
	t.Helper()
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
	return &attestation.AttestResult{Report: report}, quoteV4
}

func TestTDX_Interfaces(t *testing.T) {
	t.Run("Attester", func(_ *testing.T) {
		var _ attestation.Attester = &attestation.TDXAttester{}
	})
	t.Run("Verifier", func(_ *testing.T) {
		var _ attestation.Verifier = &attestation.TDXVerifier{}
	})
}

func TestTDXAttester_Attest(t *testing.T) {
	t.Run("error - user data too long", func(t *testing.T) {
		// given
		attester, err := attestation.NewTDXAttester()
		require.NoError(t, err)
		userData := make([]byte, attestation.IntelTdxMaxUserDataSize+1)

		// when
		_, err = attester.Attest(attestation.WithAttestUserData(userData))

		// then
		require.ErrorIs(t, err, attestation.ErrAttesterUserDataTooLong)
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
			attestation.WithVerifyMeasurement(measurement),
			attestation.WithVerifyTimestamp(timestamp),
		)

		// then
		require.NoError(t, err)
		assert.Contains(t, string(got.UserData), string(want))
	})

	t.Run("error - invalid report", func(t *testing.T) {
		// given
		report := &attestation.AttestResult{Report: []byte("invalid attestation report")}

		verifier, err := attestation.NewTDXVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifier)
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
		_, err = verifier.Verify(report, attestation.WithVerifyTimestamp(timestamp))

		// then
		require.ErrorIs(t, err, attestation.ErrVerifier)
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
			attestation.WithVerifyMeasurement(measurement),
			attestation.WithVerifyTimestamp(timestamp),
		)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifierMeasurement)
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
			attestation.WithVerifyDebug(true),
			attestation.WithVerifyMeasurement(measurement),
			attestation.WithVerifyTimestamp(timestamp),
		)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifierDebugMode)
		assert.ErrorContains(t, err, "mode mismatch")
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
		require.NoError(t, err)
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
		require.NoError(t, err)
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
		require.ErrorIs(t, err, attestation.ErrVerifierMeasurement)
		assert.ErrorContains(t, err, "unmarshaling measurement")
	})

	mismatchTestCases := []struct {
		name        string
		modifyQuote func(*pb.QuoteV4)
		wantErr     string
	}{
		{
			name: "error - tee tcb svn mismatch",
			modifyQuote: func(quoteV4 *pb.QuoteV4) {
				quoteV4.TdQuoteBody.TeeTcbSvn = []byte("wrong tee tcb svn")
			},
			wantErr: "tee tcb svn mismatch",
		},
		{
			name: "error - mr seam mismatch",
			modifyQuote: func(quoteV4 *pb.QuoteV4) {
				quoteV4.TdQuoteBody.MrSeam = []byte("wrong mr seam")
			},
			wantErr: "mr seam mismatch",
		},
		{
			name: "error - mr signer seam mismatch",
			modifyQuote: func(quoteV4 *pb.QuoteV4) {
				quoteV4.TdQuoteBody.MrSignerSeam = []byte("wrong mr signer seam")
			},
			wantErr: "mr signer seam mismatch",
		},
		{
			name: "error - seam attributes mismatch",
			modifyQuote: func(quoteV4 *pb.QuoteV4) {
				quoteV4.TdQuoteBody.SeamAttributes = []byte("wrong seam attributes")
			},
			wantErr: "seam attributes mismatch",
		},
		{
			name: "error - td attributes mismatch",
			modifyQuote: func(quoteV4 *pb.QuoteV4) {
				quoteV4.TdQuoteBody.TdAttributes = []byte("wrong td attributes")
			},
			wantErr: "td attributes mismatch",
		},
		{
			name: "error - xfam mismatch",
			modifyQuote: func(quoteV4 *pb.QuoteV4) {
				quoteV4.TdQuoteBody.Xfam = []byte("wrong xfam")
			},
			wantErr: "xfam mismatch",
		},
		{
			name: "error - mr td mismatch",
			modifyQuote: func(quoteV4 *pb.QuoteV4) {
				quoteV4.TdQuoteBody.MrTd = []byte("wrong mr td")
			},
			wantErr: "mr td mismatch",
		},
		{
			name: "error - mr config id mismatch",
			modifyQuote: func(quoteV4 *pb.QuoteV4) {
				quoteV4.TdQuoteBody.MrConfigId = []byte("wrong mr config id")
			},
			wantErr: "mr config id mismatch",
		},
		{
			name: "error - mr owner mismatch",
			modifyQuote: func(quoteV4 *pb.QuoteV4) {
				quoteV4.TdQuoteBody.MrOwner = []byte("wrong mr owner")
			},
			wantErr: "mr owner mismatch",
		},
		{
			name: "error - mr owner config mismatch",
			modifyQuote: func(quoteV4 *pb.QuoteV4) {
				quoteV4.TdQuoteBody.MrOwnerConfig = []byte("wrong mr owner config")
			},
			wantErr: "mr owner config mismatch",
		},
		{
			name: "error - missing rtmrs (quote)",
			modifyQuote: func(quoteV4 *pb.QuoteV4) {
				quoteV4.TdQuoteBody.Rtmrs = [][]byte{}
			},
			wantErr: "missing rtmrs (quote)",
		},
		{
			name: "error - rtmrs[0] mismatch",
			modifyQuote: func(quoteV4 *pb.QuoteV4) {
				quoteV4.TdQuoteBody.Rtmrs[0] = []byte("wrong rtmrs[0]")
			},
			wantErr: "rtmrs[0] mismatch",
		},
		{
			name: "error - rtmrs[1] mismatch",
			modifyQuote: func(quoteV4 *pb.QuoteV4) {
				quoteV4.TdQuoteBody.Rtmrs[1] = []byte("wrong rtmrs[1]")
			},
			wantErr: "rtmrs[1] mismatch",
		},
		{
			name: "error - rtmrs[2] mismatch",
			modifyQuote: func(quoteV4 *pb.QuoteV4) {
				quoteV4.TdQuoteBody.Rtmrs[2] = []byte("wrong rtmrs[2]")
			},
			wantErr: "rtmrs[2] mismatch",
		},
		{
			name: "error - rtmrs[3] mismatch",
			modifyQuote: func(quoteV4 *pb.QuoteV4) {
				quoteV4.TdQuoteBody.Rtmrs[3] = []byte("wrong rtmrs[3]")
			},
			wantErr: "rtmrs[3] mismatch",
		},
	}
	for _, tc := range mismatchTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			measurement := tdxReportMeasurementJSON
			timestamp := time.Unix(
				tdxReportTimestampSeconds,
				tdxReportTimestampNanoseconds,
			)
			_, quoteV4 := tdxReportFromTestData(t, tdxReportB64, timestamp)
			tc.modifyQuote(quoteV4)

			// when
			err := attestation.TDXVerifyMeasurement(measurement, quoteV4.GetTdQuoteBody())

			// then
			require.ErrorIs(t, err, attestation.ErrVerifierMeasurement)
			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}
