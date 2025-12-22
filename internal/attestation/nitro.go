package attestation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/hf/nitrite"
	"github.com/hf/nsm"
	"github.com/hf/nsm/request"
)

const AwsNitroMaxUserDataSize = 1024

type NitroAttester struct{}

func NewNitroAttester() (*NitroAttester, error) {
	return &NitroAttester{}, nil
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

	session, err := nsm.OpenDefaultSession()
	if err != nil {
		return nil, attesterError("opening nsm session", err)
	}
	defer session.Close()

	resp, err := session.Send(&request.Attestation{
		Nonce:     opts.Nonce,
		PublicKey: nil,
		UserData:  opts.UserData,
	})
	switch {
	case err != nil:
		return nil, attesterError("sending attestation request", err)
	case resp.Error != "":
		msg := fmt.Sprintf("attestation error: %s", resp.Error)
		return nil, attesterError(msg, nil)
	case resp.Attestation == nil:
		return nil, attesterError("attestation response missing attestation", nil)
	case resp.Attestation.Document == nil:
		return nil, attesterError("attestation response missing document", nil)
	}
	return &AttestResult{Report: resp.Attestation.Document}, nil
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
		UserData:  result.Document.UserData,
		PublicKey: result.Document.PublicKey,
	}
	return verifyResult, nil
}

func NitroIsDebugEnabled(document *nitrite.Document) (bool, error) {
	if len(document.PCRs) < 1 {
		return false, verifierErrorDebugMode("no pcrs provided", nil)
	}
	for i := range 3 {
		pcr, ok := document.PCRs[uint(i)]
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
