package attestation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/google/go-tdx-guest/abi"
	"github.com/google/go-tdx-guest/client"
	pb "github.com/google/go-tdx-guest/proto/tdx"
	"github.com/google/go-tdx-guest/verify"
)

const (
	IntelTdxRmrsLength      = 4
	IntelTdxMaxUserDataSize = 64
)

type TDXAttester struct{}

func NewTDXAttester() (*TDXAttester, error) {
	return &TDXAttester{}, nil
}

func (n *TDXAttester) Attest(options ...AttestOption) (*AttestResult, error) {
	opts := MakeDefaultAttestOptions()
	for _, opt := range options {
		opt(&opts)
	}

	if len(opts.UserData) > IntelTdxMaxUserDataSize {
		msg := fmt.Sprintf(
			"user data must be %d bytes or less",
			IntelTdxMaxUserDataSize,
		)
		return nil, attesterErrorUserDataTooLong(msg, nil)
	}

	tdxQP, err := client.GetQuoteProvider()
	if err != nil {
		return nil, attesterError("getting tdx quote provider", err)
	}

	var reportData [64]byte
	if opts.UserData != nil {
		copy(reportData[:], opts.UserData)
	}
	quote, err := tdxQP.GetRawQuote(reportData)
	if err != nil {
		return nil, attesterError("getting tdx quote", err)
	}
	return &AttestResult{Report: quote}, nil
}

type TDXVerifier struct{}

func NewTDXVerifier() (*TDXVerifier, error) {
	return &TDXVerifier{}, nil
}

func (n *TDXVerifier) Verify(
	attestResult *AttestResult,
	options ...VerifyOption,
) (*VerifyResult, error) {
	opts := MakeDefaultVerifyOptions()
	for _, opt := range options {
		opt(&opts)
	}

	pbQuote, err := abi.QuoteToProto(attestResult.Report)
	if err != nil {
		return nil, verifierError("converting tdx report to proto", err)
	}

	tdxOptions := verify.DefaultOptions()
	tdxOptions.Now = opts.Timestamp
	err = verify.TdxQuote(pbQuote, tdxOptions)
	if err != nil {
		return nil, verifierError("verifying tdx report", err)
	}

	quoteV4, ok := pbQuote.(*pb.QuoteV4)
	if !ok {
		return nil, verifierError("unexpected quote type", nil)
	}

	err = TDXVerifyMeasurement(opts.Measurement, quoteV4.GetTdQuoteBody())
	if err != nil {
		return nil, err
	}

	debug, err := TDXIsDebugEnabled(quoteV4)
	switch {
	case err != nil:
		return nil, err
	case opts.Debug != debug:
		msg := fmt.Sprintf("mode mismatch: expected %t, got %t",
			opts.Debug,
			debug,
		)
		return nil, verifierErrorDebugMode(msg, nil)
	}

	verifyResult := &VerifyResult{
		UserData: quoteV4.GetTdQuoteBody().GetReportData(),
	}
	return verifyResult, nil
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

type TDXMeasurement struct {
	TEETCBSVN      []byte   `json:"tee_tcb_svn"`
	MrSeam         []byte   `json:"mr_seam"`
	MrSignerSeam   []byte   `json:"mr_signer_seam"`
	SeamAttributes []byte   `json:"seam_attributes"`
	TDAttributes   []byte   `json:"td_attributes"`
	Xfam           []byte   `json:"xfam"`
	MrTD           []byte   `json:"mr_td"`
	MrConfigID     []byte   `json:"mr_config_id"`
	MrOwner        []byte   `json:"mr_owner"`
	MrOwnerConfig  []byte   `json:"mr_owner_config"`
	RTMRs          [][]byte `json:"rtmrs"`
}

func TDXVerifyMeasurement(measurementJSON string, quoteBody *pb.TDQuoteBody) error {
	if measurementJSON == "" {
		return nil
	}

	measurement := TDXMeasurement{}
	err := json.Unmarshal([]byte(measurementJSON), &measurement)
	if err != nil {
		return verifierErrorMeasurement("unmarshaling measurement", err)
	}

	switch {
	case !bytes.Equal(measurement.TEETCBSVN, quoteBody.GetTeeTcbSvn()):
		msg := fmt.Sprintf("tee tcb svn mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.TEETCBSVN),
			base64.StdEncoding.EncodeToString(quoteBody.GetTeeTcbSvn()),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.MrSeam, quoteBody.GetMrSeam()):
		msg := fmt.Sprintf("mr seam mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.MrSeam),
			base64.StdEncoding.EncodeToString(quoteBody.GetMrSeam()),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.MrSignerSeam, quoteBody.GetMrSignerSeam()):
		msg := fmt.Sprintf("mr signer seam mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.MrSignerSeam),
			base64.StdEncoding.EncodeToString(quoteBody.GetMrSignerSeam()),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.SeamAttributes, quoteBody.GetSeamAttributes()):
		msg := fmt.Sprintf("seam attributes mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.SeamAttributes),
			base64.StdEncoding.EncodeToString(quoteBody.GetSeamAttributes()),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.TDAttributes, quoteBody.GetTdAttributes()):
		msg := fmt.Sprintf("td attributes mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.TDAttributes),
			base64.StdEncoding.EncodeToString(quoteBody.GetTdAttributes()),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.Xfam, quoteBody.GetXfam()):
		msg := fmt.Sprintf("xfam mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.Xfam),
			base64.StdEncoding.EncodeToString(quoteBody.GetXfam()),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.MrTD, quoteBody.GetMrTd()):
		msg := fmt.Sprintf("mr td mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.MrTD),
			base64.StdEncoding.EncodeToString(quoteBody.GetMrTd()),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.MrConfigID, quoteBody.GetMrConfigId()):
		msg := fmt.Sprintf("mr config id mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.MrConfigID),
			base64.StdEncoding.EncodeToString(quoteBody.GetMrConfigId()),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.MrOwner, quoteBody.GetMrOwner()):
		msg := fmt.Sprintf("mr owner mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.MrOwner),
			base64.StdEncoding.EncodeToString(quoteBody.GetMrOwner()),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.MrOwnerConfig, quoteBody.GetMrOwnerConfig()):
		msg := fmt.Sprintf("mr owner config mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.MrOwnerConfig),
			base64.StdEncoding.EncodeToString(quoteBody.GetMrOwnerConfig()),
		)
		return verifierErrorMeasurement(msg, nil)
	case len(measurement.RTMRs) != IntelTdxRmrsLength:
		msg := fmt.Sprintf("missing rtmrs (measurement): expected 4, got %d",
			len(measurement.RTMRs),
		)
		return verifierErrorMeasurement(msg, nil)
	case len(quoteBody.GetRtmrs()) != IntelTdxRmrsLength:
		msg := fmt.Sprintf("missing rtmrs (quote): expected 4, got %d",
			len(quoteBody.GetRtmrs()),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.RTMRs[0], quoteBody.GetRtmrs()[0]):
		msg := fmt.Sprintf("rtmrs[0] mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.RTMRs[0]),
			base64.StdEncoding.EncodeToString(quoteBody.GetRtmrs()[0]),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.RTMRs[1], quoteBody.GetRtmrs()[1]):
		msg := fmt.Sprintf("rtmrs[1] mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.RTMRs[1]),
			base64.StdEncoding.EncodeToString(quoteBody.GetRtmrs()[1]),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.RTMRs[2], quoteBody.GetRtmrs()[2]):
		msg := fmt.Sprintf("rtmrs[2] mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.RTMRs[2]),
			base64.StdEncoding.EncodeToString(quoteBody.GetRtmrs()[2]),
		)
		return verifierErrorMeasurement(msg, nil)
	case !bytes.Equal(measurement.RTMRs[3], quoteBody.GetRtmrs()[3]):
		msg := fmt.Sprintf("rtmrs[3] mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.RTMRs[3]),
			base64.StdEncoding.EncodeToString(quoteBody.GetRtmrs()[3]),
		)
		return verifierErrorMeasurement(msg, nil)
	}
	return nil
}
