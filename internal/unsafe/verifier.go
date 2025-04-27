package unsafe

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type Verifier struct{}

func NewVerifier() (*Verifier, error) {
	return &Verifier{}, nil
}

func (n *Verifier) Verify(attestationBytes []byte) ([]byte, error) {
	attestation := Attestation{}
	err := json.Unmarshal(attestationBytes, &attestation)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling attestation: %w", err)
	}

	userdataHash := sha256.Sum256(attestation.Userdata)
	if !ECDSAVerify(attestation.Publickey, userdataHash[:], attestation.Signature) {
		return nil, fmt.Errorf("invalid signature")
	}
	return attestation.Userdata, nil
}
