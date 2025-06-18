package attestation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hf/nitrite"
	"github.com/hf/nsm"
	"github.com/hf/nsm/request"
)

type NitroAttester struct{}

func NewNitroAttester() (*NitroAttester, error) {
	return &NitroAttester{}, nil
}

func (n *NitroAttester) Attest(userdata []byte) ([]byte, error) {
	session, err := nsm.OpenDefaultSession()
	if err != nil {
		return nil, fmt.Errorf("opening nsm session: %w", err)
	}
	defer session.Close()

	resp, err := session.Send(&request.Attestation{
		Nonce:     []byte("TODO: generate nonce"),
		UserData:  userdata,
		PublicKey: []byte("TODO: generate public key"),
	})
	switch {
	case err != nil:
		return nil, fmt.Errorf("sending attestation request: %w", err)
	case resp.Error != "":
		return nil, fmt.Errorf("attestation error: %s", resp.Error)
	case resp.Attestation == nil:
		return nil, fmt.Errorf("attestation response missing attestation")
	case resp.Attestation.Document == nil:
		return nil, fmt.Errorf("attestation response missing document")
	}
	return resp.Attestation.Document, nil
}

type NitroVerifier struct{}

func NewNitroVerifier() (*NitroVerifier, error) {
	return &NitroVerifier{}, nil
}

func (n *NitroVerifier) Verify(
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

	result, err := nitrite.Verify(
		report,
		nitrite.VerifyOptions{
			CurrentTime: opts.timestamp,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("verifying report: %w", err)
	}

	err = NitroVerifyMeasurement(opts.measurement, result.Document)
	if err != nil {
		return nil, fmt.Errorf("verifying measurement: %w", err)
	}

	debug, err := NitroIsDebugEnabled(result.Document)
	switch {
	case err != nil:
		return nil, fmt.Errorf("getting debug mode: %w", err)
	case opts.debug != debug:
		return nil, fmt.Errorf("debug mode mismatch: expected %t, got %t",
			opts.debug,
			debug,
		)
	}
	return result.Document.UserData, nil
}

func NitroIsDebugEnabled(document *nitrite.Document) (bool, error) {
	if len(document.PCRs) < 1 {
		return false, fmt.Errorf("no pcrs provided")
	}
	for i := 0; i < 3; i++ {
		pcr, ok := document.PCRs[uint(i)]
		if !ok {
			return false, fmt.Errorf("missing pcr '%d'", i)
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
		return fmt.Errorf("unmarshaling measurement: %w", err)
	}

	for i, expected := range measurement.PCRs {
		got, ok := document.PCRs[i]
		if !ok {
			return fmt.Errorf("missing pcr '%d'", i)
		}
		if !bytes.Equal(expected, got) {
			return fmt.Errorf(
				"pcr '%d' mismatch: expected '%s', got '%s'",
				i,
				base64.StdEncoding.EncodeToString(expected),
				base64.StdEncoding.EncodeToString(got),
			)
		}
	}

	switch {
	case measurement.ModuleID == "":
		return nil
	case measurement.ModuleID != document.ModuleID:
		return fmt.Errorf("module id mismatch: expected '%s', got '%s'",
			measurement.ModuleID,
			document.ModuleID,
		)
	}
	return nil
}
