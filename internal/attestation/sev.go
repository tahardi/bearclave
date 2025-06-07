package attestation

import (
	"bytes"
	"encoding/base64"
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
	attestation []byte,
	options ...VerifyOption,
) ([]byte, error) {
	opts := VerifyOptions{
		measurement: "",
		timestamp:   time.Now(),
	}
	for _, opt := range options {
		opt(&opts)
	}

	pbAttestation, err := abi.ReportCertsToProto(attestation)
	if err != nil {
		return nil, fmt.Errorf("converting sev attestation to proto: %w", err)
	}

	snpOptions := verify.DefaultOptions()
	snpOptions.Now = opts.timestamp
	err = verify.SnpAttestation(pbAttestation, snpOptions)
	if err != nil {
		return nil, fmt.Errorf("verifying sev attestation: %w", err)
	}

	err = VerifyMeasurement(opts, pbAttestation.Report)
	if err != nil {
		return nil, fmt.Errorf("verifying measurement: %w", err)
	}

	return pbAttestation.Report.GetReportData(), nil
}

func VerifyMeasurement(options VerifyOptions, report *sevsnp.Report) error {
	if options.measurement == "" {
		return nil
	}

	measurement, err := base64.StdEncoding.DecodeString(options.measurement)
	if err != nil {
		return fmt.Errorf("decoding measurement: %w", err)
	}

	if !bytes.Equal(measurement, report.GetMeasurement()) {
		return fmt.Errorf(
			"measurement mismatch: expected '%x' got '%x'",
			measurement,
			report.GetMeasurement(),
		)
	}
	return nil
}
