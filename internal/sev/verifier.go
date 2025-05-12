package sev

import (
	"fmt"

	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/verify"
)

type Verifier struct{}

func NewVerifier() (*Verifier, error) {
	return &Verifier{}, nil
}

// Only annoying thing is that it always returns a 64 byte slice, even if the
// userdata is less than 64 bytes.
func (n *Verifier) Verify(attestation []byte) ([]byte, error) {
	pbAttestation, err := abi.ReportCertsToProto(attestation)
	if err != nil {
		return nil, fmt.Errorf("converting sev attestation to proto: %w", err)
	}

	err = verify.SnpAttestation(pbAttestation, verify.DefaultOptions())
	if err != nil {
		return nil, fmt.Errorf("verifying sev attestation: %w", err)
	}
	return pbAttestation.Report.GetReportData(), nil
}
