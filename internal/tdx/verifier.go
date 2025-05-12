package tdx

import (
	"fmt"

	"github.com/google/go-tdx-guest/abi"
	pb "github.com/google/go-tdx-guest/proto/tdx"
	"github.com/google/go-tdx-guest/verify"
)

type Verifier struct{}

func NewVerifier() (*Verifier, error) {
	return &Verifier{}, nil
}

func (n *Verifier) Verify(attestation []byte) ([]byte, error) {
	pbAttestation, err := abi.QuoteToProto(attestation)
	if err != nil {
		return nil, fmt.Errorf("converting tdx attestation to proto: %w", err)
	}

	err = verify.TdxQuote(pbAttestation, verify.DefaultOptions())
	if err != nil {
		return nil, fmt.Errorf("verifying tdx attestation: %w", err)
	}

	quoteV4, ok := pbAttestation.(*pb.QuoteV4)
	if !ok {
		return nil, fmt.Errorf("unexpected quote type")
	}
	return quoteV4.GetTdQuoteBody().GetReportData(), nil
}
