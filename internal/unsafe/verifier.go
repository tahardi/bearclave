package unsafe

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
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

func ECDSAVerify(publicKey []byte, data []byte, signature []byte) bool {
	r := big.Int{}
	s := big.Int{}
	sigLen := len(signature)
	r.SetBytes(signature[:(sigLen / 2)])
	s.SetBytes(signature[(sigLen / 2):])

	x := big.Int{}
	y := big.Int{}
	keyLen := len(publicKey)
	x.SetBytes(publicKey[:(keyLen / 2)])
	y.SetBytes(publicKey[(keyLen / 2):])

	rawPubKey := ecdsa.PublicKey{Curve: elliptic.P256(), X: &x, Y: &y}
	return ecdsa.Verify(&rawPubKey, data, &r, &s)
}
