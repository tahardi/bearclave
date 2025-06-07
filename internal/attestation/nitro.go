package attestation

import (
	"bytes"
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

	// TODO: Look into this resp struct and determine if we need to check/return
	// more than just the document
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
	attestation []byte,
	options ...VerifyOption,
) ([]byte, error) {
	opts := VerifyOptions{
		measurement: "",
		timestamp:   time.Now(),
	}
	for _, opt := range options {
		opt(&opts)
	}

	result, err := nitrite.Verify(
		attestation,
		nitrite.VerifyOptions{
			CurrentTime: opts.timestamp,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("verifying attestation: %w", err)
	}

	err = VerifyPCRs(opts, result.Document)
	if err != nil {
		return nil, fmt.Errorf("verifying pcrs: %w", err)
	}

	return result.Document.UserData, nil
}

func VerifyPCRs(options VerifyOptions, document *nitrite.Document) error {
	if options.measurement == "" {
		return nil
	}

	var expected map[uint][]byte
	err := json.Unmarshal([]byte(options.measurement), &expected)
	if err != nil {
		return fmt.Errorf("unmarshaling pcrs: %w", err)
	}

	for pcr, expectedVal := range expected {
		gotVal, ok := document.PCRs[pcr]
		if !ok {
			return fmt.Errorf("missing pcr %d", pcr)
		}
		if !bytes.Equal(expectedVal, gotVal) {
			return fmt.Errorf(
				"pcr %d mismatch: expected '%x', got '%x'",
				pcr,
				expectedVal,
				gotVal,
			)
		}
	}
	return nil
}
