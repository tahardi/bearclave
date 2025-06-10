package attestation

import (
	"bytes"
	"encoding/hex"
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
		debug:       false,
		measurement: "",
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

	err = TDXVerifyMeasurement(opts.measurement, quoteV4.GetTdQuoteBody())
	if err != nil {
		return nil, fmt.Errorf("verifying measurement: %w", err)
	}

	debug, err := TDXIsDebugEnabled(quoteV4)
	switch {
	case err != nil:
		return nil, fmt.Errorf("getting debug mode: %w", err)
	case opts.debug != debug:
		return nil, fmt.Errorf("debug mode mismatch: expected %t, got %t",
			opts.debug,
			debug,
		)
	}
	return quoteV4.GetTdQuoteBody().GetReportData(), nil
}

func TDXIsDebugEnabled(quoteV4 *pb.QuoteV4) (bool, error) {
	tdAttributes := quoteV4.GetTdQuoteBody().GetTdAttributes()

	// Documentation states that if any of bits 7:0 are set to 1, then
	// the TD is in debug mode. Thus, if they are all 0, debug is not enabled.
	// Also, the documentation states that all fields are little endian
	// https://download.01.org/intel-sgx/latest/dcap-latest/linux/docs/Intel_TDX_DCAP_Quoting_Library_API.pdf
	if tdAttributes[0]&0xFF == 0 {
		return false, nil
	}
	return true, nil
}

func TDXVerifyMeasurement(measurement string, quoteBody *pb.TDQuoteBody) error {
	if measurement == "" {
		return nil
	}

	expected, err := TDXParseMeasurement(measurement)
	if err != nil {
		return fmt.Errorf("parsing measurement: %w", err)
	}

	got := quoteBody.GetMrTd()
	if !bytes.Equal(expected, got) {
		return fmt.Errorf(
			"measurement mismatch: expected '%x' got '%x'",
			expected,
			got,
		)
	}
	return nil
}

func TDXParseMeasurement(measurement string) ([]byte, error) {
	return hex.DecodeString(measurement)
}
