package attestation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
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
	quote, err := tdxQP.GetRawQuote(reportData)
	if err != nil {
		return nil, fmt.Errorf("getting tdx quote: %w", err)
	}
	return quote, nil
}

type TDXVerifier struct{}

func NewTDXVerifier() (*TDXVerifier, error) {
	return &TDXVerifier{}, nil
}

func (n *TDXVerifier) Verify(
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

	pbQuote, err := abi.QuoteToProto(report)
	if err != nil {
		return nil, fmt.Errorf("converting tdx report to proto: %w", err)
	}

	tdxOptions := verify.DefaultOptions()
	tdxOptions.Now = opts.timestamp
	err = verify.TdxQuote(pbQuote, tdxOptions)
	if err != nil {
		return nil, fmt.Errorf("verifying tdx report: %w", err)
	}

	quoteV4, ok := pbQuote.(*pb.QuoteV4)
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
		return fmt.Errorf("unmarshaling measurement: %w", err)
	}

	switch {
	case !bytes.Equal(measurement.TEETCBSVN, quoteBody.GetTeeTcbSvn()):
		return fmt.Errorf("tee tcb svn mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.TEETCBSVN),
			base64.StdEncoding.EncodeToString(quoteBody.GetTeeTcbSvn()),
		)
	case !bytes.Equal(measurement.MrSeam, quoteBody.GetMrSeam()):
		return fmt.Errorf("mr seam mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.MrSeam),
			base64.StdEncoding.EncodeToString(quoteBody.GetMrSeam()),
		)
	case !bytes.Equal(measurement.MrSignerSeam, quoteBody.GetMrSignerSeam()):
		return fmt.Errorf("mr signer seam mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.MrSignerSeam),
			base64.StdEncoding.EncodeToString(quoteBody.GetMrSignerSeam()),
		)
	case !bytes.Equal(measurement.SeamAttributes, quoteBody.GetSeamAttributes()):
		return fmt.Errorf("seam attributes mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.SeamAttributes),
			base64.StdEncoding.EncodeToString(quoteBody.GetSeamAttributes()),
		)
	case !bytes.Equal(measurement.TDAttributes, quoteBody.GetTdAttributes()):
		return fmt.Errorf("td attributes mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.TDAttributes),
			base64.StdEncoding.EncodeToString(quoteBody.GetTdAttributes()),
		)
	case !bytes.Equal(measurement.Xfam, quoteBody.GetXfam()):
		return fmt.Errorf("xfam mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.Xfam),
			base64.StdEncoding.EncodeToString(quoteBody.GetXfam()),
		)
	case !bytes.Equal(measurement.MrTD, quoteBody.GetMrTd()):
		return fmt.Errorf("mr td mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.MrTD),
			base64.StdEncoding.EncodeToString(quoteBody.GetMrTd()),
		)
	case !bytes.Equal(measurement.MrConfigID, quoteBody.GetMrConfigId()):
		return fmt.Errorf("mr config id mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.MrConfigID),
			base64.StdEncoding.EncodeToString(quoteBody.GetMrConfigId()),
		)
	case !bytes.Equal(measurement.MrOwner, quoteBody.GetMrOwner()):
		return fmt.Errorf("mr owner mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.MrOwner),
			base64.StdEncoding.EncodeToString(quoteBody.GetMrOwner()),
		)
	case !bytes.Equal(measurement.MrOwnerConfig, quoteBody.GetMrOwnerConfig()):
		return fmt.Errorf("mr owner config mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.MrOwnerConfig),
			base64.StdEncoding.EncodeToString(quoteBody.GetMrOwnerConfig()),
		)
	case len(measurement.RTMRs) != 4:
		return fmt.Errorf("missing rtmrs (measurement): expected 4, got %d",
			len(measurement.RTMRs),
		)
	case len(quoteBody.GetRtmrs()) != 4:
		return fmt.Errorf("missing rtmrs (quote): expected 4, got %d",
			len(quoteBody.GetRtmrs()),
		)
	case !bytes.Equal(measurement.RTMRs[0], quoteBody.GetRtmrs()[0]):
		return fmt.Errorf("rtmrs[0] mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.RTMRs[0]),
			base64.StdEncoding.EncodeToString(quoteBody.GetRtmrs()[0]),
		)
	case !bytes.Equal(measurement.RTMRs[1], quoteBody.GetRtmrs()[1]):
		return fmt.Errorf("rtmrs[1] mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.RTMRs[1]),
			base64.StdEncoding.EncodeToString(quoteBody.GetRtmrs()[1]),
		)
	case !bytes.Equal(measurement.RTMRs[2], quoteBody.GetRtmrs()[2]):
		return fmt.Errorf("rtmrs[2] mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.RTMRs[2]),
			base64.StdEncoding.EncodeToString(quoteBody.GetRtmrs()[2]),
		)
	case !bytes.Equal(measurement.RTMRs[3], quoteBody.GetRtmrs()[3]):
		return fmt.Errorf("rtmrs[3] mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(measurement.RTMRs[3]),
			base64.StdEncoding.EncodeToString(quoteBody.GetRtmrs()[3]),
		)
	}
	return nil
}
