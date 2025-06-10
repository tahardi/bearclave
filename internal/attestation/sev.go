package attestation

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/client"
	"github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/google/go-sev-guest/verify"
)

const AMD_SEV_USERDATA_SIZE = 64

type SEVAttester struct{}

func NewSEVAttester() (*SEVAttester, error) {
	return &SEVAttester{}, nil
}

func (n *SEVAttester) Attest(userdata []byte) ([]byte, error) {
	if len(userdata) > AMD_SEV_USERDATA_SIZE {
		return nil, fmt.Errorf(
			"userdata must be less than %d bytes",
			AMD_SEV_USERDATA_SIZE,
		)
	}

	sevQP, err := client.GetQuoteProvider()
	if err != nil {
		return nil, fmt.Errorf("getting sev quote provider: %w", err)
	}

	var reportData [64]byte
	copy(reportData[:], userdata)
	attestation, err := sevQP.GetRawQuote(reportData)
	if err != nil {
		return nil, fmt.Errorf("getting sev quote: %w", err)
	}
	return attestation, nil
}

type SEVVerifier struct{}

func NewSEVVerifier() (*SEVVerifier, error) {
	return &SEVVerifier{}, nil
}

// Only annoying thing is that it always returns a 64 byte slice, even if the
// userdata is less than 64 bytes.
func (n *SEVVerifier) Verify(
	report []byte,
	options ...VerifyOption,
) ([]byte, error) {
	opts := VerifyOptions{
		debug:       false,
		measurement: "",
		timestamp:   time.Now(),
	}
	for _, opt := range options {
		opt(&opts)
	}

	pbReport, err := abi.ReportCertsToProto(report)
	if err != nil {
		return nil, fmt.Errorf("converting sev report to proto: %w", err)
	}

	snpOptions := verify.DefaultOptions()
	snpOptions.Now = opts.timestamp
	err = verify.SnpAttestation(pbReport, snpOptions)
	if err != nil {
		return nil, fmt.Errorf("verifying sev report: %w", err)
	}

	err = SEVVerifyMeasurement(opts.measurement, pbReport.Report)
	if err != nil {
		return nil, fmt.Errorf("verifying measurement: %w", err)
	}

	debug, err := SEVIsDebugEnabled(pbReport.Report)
	switch {
	case err != nil:
		return nil, fmt.Errorf("getting debug mode: %w", err)
	case opts.debug != debug:
		return nil, fmt.Errorf("debug mode mismatch: expected %t, got %t",
			opts.debug,
			debug,
		)
	}
	return pbReport.Report.GetReportData(), nil
}

func SEVIsDebugEnabled(report *sevsnp.Report) (bool, error) {
	policy, err := abi.ParseSnpPolicy(report.GetPolicy())
	if err != nil {
		return false, fmt.Errorf("parsing policy: %w", err)
	}
	return policy.Debug, nil
}

func SEVVerifyMeasurement(measurement string, report *sevsnp.Report) error {
	if measurement == "" {
		return nil
	}

	expected, err := SEVParseMeasurement(measurement)
	if err != nil {
		return fmt.Errorf("parsing measurement: %w", err)
	}

	got := report.GetMeasurement()
	if !bytes.Equal(expected, got) {
		return fmt.Errorf(
			"measurement mismatch: expected '%x' got '%x'",
			expected,
			got,
		)
	}
	return nil
}

func SEVParseMeasurement(measurement string) ([]byte, error) {
	return hex.DecodeString(measurement)
}
