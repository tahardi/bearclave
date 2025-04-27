package unsafe

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type Attester struct {
	privateKey *ecdsa.PrivateKey
	publicKey  []byte
}

func NewAttester() (*Attester, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating private key: %w", err)
	}
	return NewAttesterWithPrivateKey(privateKey)
}

func NewAttesterWithPrivateKey(privateKey *ecdsa.PrivateKey) (*Attester, error) {
	publicKey := append(
		privateKey.PublicKey.X.Bytes(),
		privateKey.PublicKey.Y.Bytes()...,
	)
	return &Attester{privateKey: privateKey, publicKey: publicKey}, nil
}

type Attestation struct {
	Userdata  []byte `json:"userdata"`
	Signature []byte `json:"signature"`
	Publickey []byte `json:"publickey"`
}

func (a *Attester) Attest(userdata []byte) ([]byte, error) {
	userdataHash := sha256.Sum256(userdata)
	signature, err := ECDSASign(a.privateKey, userdataHash[:])
	if err != nil {
		return nil, fmt.Errorf("signing userdata: %w", err)
	}

	attestation := Attestation{
		Userdata:  userdata,
		Signature: signature,
		Publickey: a.publicKey,
	}
	attestationBytes, err := json.Marshal(attestation)
	if err != nil {
		return nil, fmt.Errorf("marshaling attestation: %w", err)
	}
	return attestationBytes, nil
}
