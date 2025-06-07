package attestation

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

const (
	NoTEEMeasurement    = "Not a TEE platform. Code measurements are not real."
	NoTEEValidityPeriod = int64(31536000)
)

type NoTEEAttester struct {
	privateKey *ecdsa.PrivateKey
	publicKey  []byte
}

func NewNoTEEAttester() (*NoTEEAttester, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating private key: %w", err)
	}
	return NewNoTEEAttesterWithPrivateKey(privateKey)
}

func NewNoTEEAttesterWithPrivateKey(
	privateKey *ecdsa.PrivateKey,
) (*NoTEEAttester, error) {
	publicKey := append(
		privateKey.X.Bytes(),
		privateKey.Y.Bytes()...,
	)
	return &NoTEEAttester{privateKey: privateKey, publicKey: publicKey}, nil
}

type Report struct {
	Userdata    []byte `json:"userdata"`
	Signature   []byte `json:"signature"`
	Publickey   []byte `json:"publickey"`
	Timestamp   int64  `json:"timestamp"`
	Measurement string `json:"measurement"`
}

func (a *NoTEEAttester) Attest(userdata []byte) ([]byte, error) {
	userdataHash := sha256.Sum256(userdata)
	signature, err := ECDSASign(a.privateKey, userdataHash[:])
	if err != nil {
		return nil, fmt.Errorf("signing userdata: %w", err)
	}

	report := Report{
		Userdata:    userdata,
		Signature:   signature,
		Publickey:   a.publicKey,
		Timestamp:   time.Now().Unix(),
		Measurement: NoTEEMeasurement,
	}
	reportBytes, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("marshaling report: %w", err)
	}
	return reportBytes, nil
}

type NoTEEVerifier struct{}

func NewNoTEEVerifier() (*NoTEEVerifier, error) {
	return &NoTEEVerifier{}, nil
}

func (n *NoTEEVerifier) Verify(
	reportBytes []byte,
	options ...VerifyOption,
) ([]byte, error) {
	opts := VerifyOptions{
		measurement: "",
		timestamp:   time.Now(),
	}
	for _, opt := range options {
		opt(&opts)
	}

	report := Report{}
	err := json.Unmarshal(reportBytes, &report)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling report: %w", err)
	}

	userdataHash := sha256.Sum256(report.Userdata)
	if !ECDSAVerify(report.Publickey, userdataHash[:], report.Signature) {
		return nil, fmt.Errorf("invalid signature")
	}

	if opts.timestamp.Unix() < report.Timestamp ||
		opts.timestamp.Unix() > report.Timestamp+NoTEEValidityPeriod {
		return nil, fmt.Errorf("certificate has expired or is not yet valid")
	}

	if opts.measurement != "" && opts.measurement != report.Measurement {
		return nil, fmt.Errorf(
			"measurement mismatch: expected '%s' got '%s'",
			opts.measurement,
			report.Measurement,
		)
	}

	return report.Userdata, nil
}

func ECDSASign(privateKey *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	r, s, err := ecdsa.Sign(crand.Reader, privateKey, data)
	if err != nil {
		return nil, fmt.Errorf("ecdsa signing data: %w", err)
	}
	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
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
