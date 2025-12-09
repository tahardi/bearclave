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
  "version": 5,
  "guest_svn": 0,
  "policy": 196608,
  "family_id": "AAAAAAAAAAAAAAAAAAAAAA==",
  "image_id": "AAAAAAAAAAAAAAAAAAAAAA==",
  "vmpl": 0,
  "current_tcb": 16004385700791189508,
  "platform_info": 37,
  "signer_info": 0,
  "measurement": "t0fVVFLguekHl3Cknjl8Xm2Vc1geJG2nuqxPKLXNxbG20ZJRuO5gD9FqNwj1hAbz",
  "host_data": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "id_key_digest": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
  "author_key_digest": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
  "report_id": "mQCG/KnohXCpN98FyZ5vQ2JknQdZ2AhhsnOJbOLp58U=",
  "report_id_ma": "//////////////////////////////////////////8=",
  "reported_tcb": 16004385700791189508,
  "chip_id": "BGnXVODUH+qJbpsQKqxKve5nZ2TmfU4yuw4baIU9u7v/MdJmOLCc7wFFxvsFLDggUNohMViMUOBfeqsK480+AQ==",
  "committed_tcb": 16004385700791189508,
  "current_build": 35,
  "current_minor": 55,
  "current_major": 1,
  "committed_build": 35,
  "committed_minor": 55,
  "committed_major": 1,
  "launch_tcb": 16004385700791189508,
  "cpuid_1eax_fms": 10489617
}`
	// The testdata SEV report was generated on Fri Dec 05 2025 03:02:30 GMT+0000
	sevReportTimestampSeconds = int64(1764903750)
)

func sevReportFromTestData(
	t *testing.T,
	reportB64 string,
	timestamp time.Time,
) (*attestation.AttestResult, *sevsnp.Report) {
	report, err := base64.StdEncoding.DecodeString(reportB64)
	require.NoError(t, err)

	pbReport, err := abi.ReportCertsToProto(report)
	require.NoError(t, err)

	opts := verify.DefaultOptions()
	opts.Now = timestamp
	err = verify.SnpAttestation(pbReport, opts)
	require.NoError(t, err)

	return &attestation.AttestResult{Report: report}, pbReport.GetReport()
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
		timestamp := time.Unix(sevReportTimestampSeconds, 0)
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
		assert.Contains(t, string(got.UserData), string(want))
	})

	t.Run("error - invalid report", func(t *testing.T) {
		// given
		report := &attestation.AttestResult{Report: []byte("invalid attestation report")}

		verifier, err := attestation.NewSEVVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report)

		// then
		assert.ErrorContains(t, err, "converting sev report to proto")
	})

	t.Run("error - expired report", func(t *testing.T) {
		// given
		timestamp := time.Unix(sevReportTimestampSeconds, 0)
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
		timestamp := time.Unix(sevReportTimestampSeconds, 0)
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
		timestamp := time.Unix(sevReportTimestampSeconds, 0)
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
		timestamp := time.Unix(sevReportTimestampSeconds, 0)
		_, sevReport := sevReportFromTestData(t, sevReportB64, timestamp)

		// According to Pg. 31 of the SEV ABI Specification document, the 19th
		// bit of the Policy denotes whether debugging is enabled or not.
		// https://www.amd.com/content/dam/amd/en/documents/epyc-technical-docs/specifications/56860.pdf
		sevReport.Policy |= uint64(1 << 19)

		// when
		got, err := attestation.SEVIsDebugEnabled(sevReport)

		// then
		assert.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("happy path - debug disabled", func(t *testing.T) {
		// given
		timestamp := time.Unix(sevReportTimestampSeconds, 0)
		_, sevReport := sevReportFromTestData(t, sevReportB64, timestamp)

		// when
		got, err := attestation.SEVIsDebugEnabled(sevReport)

		// then
		assert.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("error - parsing policy", func(t *testing.T) {
		// given
		timestamp := time.Unix(sevReportTimestampSeconds, 0)
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
		timestamp := time.Unix(sevReportTimestampSeconds, 0)
		_, sevReport := sevReportFromTestData(t, sevReportB64, timestamp)

		// when
		err := attestation.SEVVerifyMeasurement(measurement, sevReport)

		// then
		assert.NoError(t, err)
	})

	t.Run("happy path - no measurement", func(t *testing.T) {
		// given
		measurement := ""
		timestamp := time.Unix(sevReportTimestampSeconds, 0)
		_, sevReport := sevReportFromTestData(t, sevReportB64, timestamp)

		// when
		err := attestation.SEVVerifyMeasurement(measurement, sevReport)

		// then
		assert.NoError(t, err)
	})

	t.Run("error - invalid measurement", func(t *testing.T) {
		// given
		measurement := "invalid measurement"
		timestamp := time.Unix(sevReportTimestampSeconds, 0)
		_, sevReport := sevReportFromTestData(t, sevReportB64, timestamp)

		// when
		err := attestation.SEVVerifyMeasurement(measurement, sevReport)

		// then
		assert.ErrorContains(t, err, "unmarshaling measurement")
	})

	mismatchTestCases := []struct {
		name         string
		modifyReport func(*sevsnp.Report)
		wantErr      string
	}{
		{
			name: "error - version mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.Version += 1
			},
			wantErr: "version mismatch",
		},
		{
			name: "error - guest svn mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.GuestSvn += 1
			},
			wantErr: "guest svn mismatch",
		},
		{
			name: "error - policy mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.Policy += 1
			},
			wantErr: "policy mismatch",
		},
		{
			name: "error - family id mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.FamilyId = []byte("wrong family id")
			},
			wantErr: "family id mismatch",
		},
		{
			name: "error - image id mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.ImageId = []byte("wrong image id")
			},
			wantErr: "image id mismatch",
		},
		{
			name: "error - vmpl mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.Vmpl += 1
			},
			wantErr: "vmpl mismatch",
		},
		{
			name: "error - current tcb mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.CurrentTcb += 1
			},
			wantErr: "current tcb mismatch",
		},
		{
			name: "error - platform info mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.PlatformInfo += 1
			},
			wantErr: "platform info mismatch",
		},
		{
			name: "error - signer info mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.SignerInfo += 1
			},
			wantErr: "signer info mismatch",
		},
		{
			name: "error - measurement mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.Measurement = []byte("wrong measurement")
			},
			wantErr: "measurement mismatch",
		},
		{
			name: "error - host data mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.HostData = []byte("wrong host data")
			},
			wantErr: "host data mismatch",
		},
		{
			name: "error - id key digest mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.IdKeyDigest = []byte("wrong id key digest")
			},
			wantErr: "id key digest mismatch",
		},
		{
			name: "error - author key digest mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.AuthorKeyDigest = []byte("wrong author key digest")
			},
			wantErr: "author key digest mismatch",
		},
		{
			name: "error - report id mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.ReportId = []byte("wrong report id")
			},
			wantErr: "report id mismatch",
		},
		{
			name: "error - reported tcb mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.ReportedTcb += 1
			},
			wantErr: "reported tcb mismatch",
		},
		{
			name: "error - chip id mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.ChipId = []byte("wrong chip id")
			},
			wantErr: "chip id mismatch",
		},
		{
			name: "error - committed tcb mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.CommittedTcb += 1
			},
			wantErr: "committed tcb mismatch",
		},
		{
			name: "error - current build mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.CurrentBuild += 1
			},
			wantErr: "current build mismatch",
		},
		{
			name: "error - current minor mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.CurrentMinor += 1
			},
			wantErr: "current minor mismatch",
		},
		{
			name: "error - current major mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.CurrentMajor += 1
			},
			wantErr: "current major mismatch",
		},
		{
			name: "error - committed build mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.CommittedBuild += 1
			},
			wantErr: "committed build mismatch",
		},
		{
			name: "error - committed minor mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.CommittedMinor += 1
			},
			wantErr: "committed minor mismatch",
		},
		{
			name: "error - committed major mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.CommittedMajor += 1
			},
			wantErr: "committed major mismatch",
		},
		{
			name: "error - launch tcb mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.LaunchTcb += 1
			},
			wantErr: "launch tcb mismatch",
		},
		{
			name: "error - cpuid 1eax fms mismatch",
			modifyReport: func(report *sevsnp.Report) {
				report.Cpuid1EaxFms += 1
			},
			wantErr: "cpuid 1eax fms mismatch",
		},
	}

	for _, tc := range mismatchTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			measurement := sevReportMeasurementJSON
			timestamp := time.Unix(sevReportTimestampSeconds, 0)
			_, sevReport := sevReportFromTestData(t, sevReportB64, timestamp)
			tc.modifyReport(sevReport)

			// when
			err := attestation.SEVVerifyMeasurement(measurement, sevReport)

			// then
			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}
