package attestation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/hf/nitrite"
	"github.com/tahardi/bearclave/internal/drivers"
)

const (
	AwsNitroMaxUserDataSize = 1024
	AWSNitroDebugPCRRange   = uint(3)
)

type NitroAttester struct {
	client drivers.NSM
}

func NewNitroAttester() (*NitroAttester, error) {
	client, err := drivers.NewNSMClient()
	if err != nil {
		return nil, attesterError("making nsm client", err)
	}
	return NewNitroAttesterWithClient(client)
}

func NewNitroAttesterWithClient(client drivers.NSM) (*NitroAttester, error) {
	return &NitroAttester{client: client}, nil
}

func (n *NitroAttester) Close() error {
	return n.client.Close()
}

func (n *NitroAttester) Attest(options ...AttestOption) (*AttestResult, error) {
	opts := MakeDefaultAttestOptions()
	for _, opt := range options {
		opt(&opts)
	}

	if len(opts.UserData) > AwsNitroMaxUserDataSize {
		msg := fmt.Sprintf(
			"user data must be %d bytes or less",
			AwsNitroMaxUserDataSize,
		)
		return nil, attesterErrorUserData(msg, nil)
	}

	attestation, err := n.client.GetAttestation(
		opts.Nonce,
		nil,
		opts.UserData,
	)
	if err != nil {
		return nil, attesterError("getting attestation", err)
	}
	return &AttestResult{Report: attestation}, nil
}

type NitroVerifier struct{}

func NewNitroVerifier() (*NitroVerifier, error) {
	return &NitroVerifier{}, nil
}

func (n *NitroVerifier) Verify(
	attestResult *AttestResult,
	options ...VerifyOption,
) (*VerifyResult, error) {
	opts := MakeDefaultVerifyOptions()
	for _, opt := range options {
		opt(&opts)
	}

	result, err := nitrite.Verify(
		attestResult.Report,
		nitrite.VerifyOptions{
			CurrentTime: opts.Timestamp,
		},
	)
	if err != nil {
		return nil, verifierError("verifying report", err)
	}

	err = NitroVerifyMeasurement(opts.Measurement, result.Document)
	if err != nil {
		return nil, err
	}

	err = NitroVerifyNonce(opts.Nonce, result.Document)
	if err != nil {
		return nil, err
	}

	debug, err := NitroIsDebugEnabled(result.Document)
	switch {
	case err != nil:
		return nil, err
	case opts.Debug != debug:
		msg := fmt.Sprintf("mode mismatch: expected %t got %t", opts.Debug, debug)
		return nil, verifierErrorDebugMode(msg, nil)
	}

	verifyResult := &VerifyResult{
		UserData: result.Document.UserData,
	}
	return verifyResult, nil
}

func NitroIsDebugEnabled(document *nitrite.Document) (bool, error) {
	if len(document.PCRs) < 1 {
		return false, verifierErrorDebugMode("no pcrs provided", nil)
	}
	for i := range AWSNitroDebugPCRRange {
		pcr, ok := document.PCRs[i]
		if !ok {
			msg := fmt.Sprintf("missing pcr '%d'", i)
			return false, verifierErrorDebugMode(msg, nil)
		}
		for _, b := range pcr {
			if b != 0 {
				return false, nil
			}
		}
	}
	return true, nil
}

type NitroMeasurement struct {
	PCRs     map[uint][]byte `json:"pcrs"`
	ModuleID string          `json:"module_id"`
}

func NitroVerifyMeasurement(measurementJSON string, document *nitrite.Document) error {
	if measurementJSON == "" {
		return nil
	}

	measurement := NitroMeasurement{}
	err := json.Unmarshal([]byte(measurementJSON), &measurement)
	if err != nil {
		return verifierErrorMeasurement("unmarshaling measurement", err)
	}

	for i, expected := range measurement.PCRs {
		got, ok := document.PCRs[i]
		if !ok {
			msg := fmt.Sprintf("missing pcr '%d'", i)
			return verifierErrorMeasurement(msg, nil)
		}
		if !bytes.Equal(expected, got) {
			msg := fmt.Sprintf(
				"pcr '%d' mismatch: expected '%s', got '%s'",
				i,
				base64.StdEncoding.EncodeToString(expected),
				base64.StdEncoding.EncodeToString(got),
			)
			return verifierErrorMeasurement(msg, nil)
		}
	}

	switch {
	case measurement.ModuleID == "":
		return nil
	case measurement.ModuleID != document.ModuleID:
		msg := fmt.Sprintf("module id mismatch: expected '%s', got '%s'",
			measurement.ModuleID,
			document.ModuleID,
		)
		return verifierErrorMeasurement(msg, nil)
	}
	return nil
}

func NitroVerifyNonce(nonce []byte, document *nitrite.Document) error {
	if len(nonce) == 0 {
		return nil
	}
	if !bytes.Equal(nonce, document.Nonce) {
		msg := fmt.Sprintf("nonce mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(nonce),
			base64.StdEncoding.EncodeToString(document.Nonce),
		)
		return verifierErrorNonce(msg, nil)
	}
	return nil
}
