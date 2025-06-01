package attestation

import (
	"fmt"
	"time"

	"github.com/google/go-tdx-guest/abi"
	"github.com/google/go-tdx-guest/client"
	pb "github.com/google/go-tdx-guest/proto/tdx"
	"github.com/google/go-tdx-guest/verify"
)

const INTEL_TDX_USERDATA_SIZE = 64

type TDXAttester struct{}

func NewTDXAttester() (*TDXAttester, error) {
	return &TDXAttester{}, nil
}

func (n *TDXAttester) Attest(userdata []byte) ([]byte, error) {
	if len(userdata) > INTEL_TDX_USERDATA_SIZE {
		return nil, fmt.Errorf(
			"userdata must be less than %d bytes",
			INTEL_TDX_USERDATA_SIZE,
		)
	}

	tdxQP, err := client.GetQuoteProvider()
	if err != nil {
		return nil, fmt.Errorf("getting tdx quote provider: %w", err)
	}

	var reportData [64]byte
	copy(reportData[:], userdata)
	attestation, err := tdxQP.GetRawQuote(reportData)
	if err != nil {
		return nil, fmt.Errorf("getting tdx quote: %w", err)
	}
	return attestation, nil
}

type TDXVerifier struct{}

func NewTDXVerifier() (*TDXVerifier, error) {
	return &TDXVerifier{}, nil
}

func (n *TDXVerifier) Verify(
	attestation []byte,
	options ...VerifyOption,
) ([]byte, error) {
	opts := VerifyOptions{
		measurement: nil,
		timestamp:   time.Now(),
	}
	for _, opt := range options {
		opt(&opts)
	}

	pbAttestation, err := abi.QuoteToProto(attestation)
	if err != nil {
		return nil, fmt.Errorf("converting tdx attestation to proto: %w", err)
	}

	tdxOptions := verify.DefaultOptions()
	tdxOptions.Now = opts.timestamp
	err = verify.TdxQuote(pbAttestation, tdxOptions)
	if err != nil {
		return nil, fmt.Errorf("verifying tdx attestation: %w", err)
	}

	quoteV4, ok := pbAttestation.(*pb.QuoteV4)
	if !ok {
		return nil, fmt.Errorf("unexpected quote type")
	}
	return quoteV4.GetTdQuoteBody().GetReportData(), nil
}
