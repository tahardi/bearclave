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

func (n *Verifier) Verify(attestation []byte) ([]byte, error) {
	pbAttestation, err := abi.ReportCertsToProto(attestation)
	if err != nil {
		return nil, fmt.Errorf("converting attestation to proto: %w", err)
	}

	err = verify.SnpAttestation(pbAttestation, verify.DefaultOptions())
	if err != nil {
		return nil, fmt.Errorf("verifying attestation: %w", err)
	}
	return pbAttestation.Report.GetReportData(), nil
}
