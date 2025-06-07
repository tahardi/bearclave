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

	expectedPCRs, err := ParsePCRs(options.measurement)
	if err != nil {
		return fmt.Errorf("parsing measurement: %w", err)
	}

	for i, expectedPCR := range expectedPCRs {
		gotPCR, ok := document.PCRs[i]
		if !ok {
			return fmt.Errorf("missing pcr '%d'", i)
		}
		if !bytes.Equal(expectedPCR, gotPCR) {
			return fmt.Errorf(
				"pcr '%d' mismatch: expected '%x', got '%x'",
				i,
				expectedPCR,
				gotPCR,
			)
		}
	}
	return nil
}

func ParsePCRs(measurement string) (map[uint][]byte, error) {
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
