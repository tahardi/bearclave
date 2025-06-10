package attestation

import (
	"bytes"
	"encoding/hex"
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

// TODO: userdata limit?
// TODO: Check limits for userdata, nonce, and public key?
// TODO: Consider how to update the Attest interface to account for
// other arguments (e.g., Nonce, Publickey). Maybe it takes `any` instead
// and tries to unmarshal into platform specific struct?
func (n *NitroAttester) Attest(userdata []byte) ([]byte, error) {
	// TODO: Can I hold onto a long-lived session instance?
	// TODO: Can I mock the session for testing of this function?
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

func NitroVerifyMeasurement(measurement string, document *nitrite.Document) error {
	if measurement == "" {
		return nil
	}

	expectedPCRs, err := NitroParseMeasurement(measurement)
	if err != nil {
		return fmt.Errorf("parsing measurement: %w", err)
	}

	for i, expected := range expectedPCRs {
		got, ok := document.PCRs[i]
		if !ok {
			return fmt.Errorf("missing pcr '%d'", i)
		}
		if !bytes.Equal(expected, got) {
			return fmt.Errorf(
				"pcr '%d' mismatch: expected '%x', got '%x'",
				i,
				expected,
				got,
			)
		}
	}
	return nil
}

func NitroParseMeasurement(measurement string) (map[uint][]byte, error) {
	var pcrHexStrings map[uint]string
	err := json.Unmarshal([]byte(measurement), &pcrHexStrings)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling pcrs: %w", err)
	}

	pcrs := make(map[uint][]byte)
	for k, v := range pcrHexStrings {
		pcr, err := hex.DecodeString(v)
		if err != nil {
			return nil, fmt.Errorf("decoding hex string for PCR%v: %w", k, err)
		}
		pcrs[k] = pcr
	}

	return pcrs, nil
}
