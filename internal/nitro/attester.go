package nitro

import (
	"fmt"

	"github.com/hf/nsm"
	"github.com/hf/nsm/request"
)

type Attester struct{}

func NewAttester() (*Attester, error) {
	return &Attester{}, nil
}

// TODO: Check limits for userdata, nonce, and public key?
// TODO: Consider how to update the Attest interface to account for
// other arguments (e.g., Nonce, Publickey). Maybe it takes `any` instead
// and tries to unmarshal into platform specific struct?
func (n *Attester) Attest(userdata []byte) ([]byte, error) {
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
