package attestation

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
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
	Nonce       []byte `json:"nonce"`
	PublicKey   []byte `json:"publickey"`
	Signature   []byte `json:"signature"`
	VerifyKey   []byte `json:"verifykey"`
	Timestamp   int64  `json:"timestamp"`
	Measurement string `json:"measurement"`
}

func (a *NoTEEAttester) Attest(options ...AttestOption) (*AttestResult, error) {
	opts := AttestOptions{
		nonce:     nil,
		publicKey: nil,
		userData:  nil,
	}
	for _, opt := range options {
		opt(&opts)
	}

	signDataHash := sha256.Sum256([]byte(NoTEEMeasurement))
	signature, err := ECDSASign(a.privateKey, signDataHash[:])
	if err != nil {
		return nil, fmt.Errorf("signing userdata: %w", err)
	}

	report := Report{
		Nonce:       opts.nonce,
		PublicKey:   opts.publicKey,
		Userdata:    opts.userData,
		Signature:   signature,
		VerifyKey:   a.publicKey,
		Timestamp:   time.Now().Unix(),
		Measurement: NoTEEMeasurement,
	}
	reportBytes, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("marshaling report: %w", err)
	}
	return &AttestResult{Report: reportBytes}, nil
}

type NoTEEVerifier struct{}

func NewNoTEEVerifier() (*NoTEEVerifier, error) {
	return &NoTEEVerifier{}, nil
}

func (n *NoTEEVerifier) Verify(
	attestResult *AttestResult,
	options ...VerifyOption,
) (*VerifyResult, error) {
	opts := VerifyOptions{
		measurement: "",
		timestamp:   time.Now(),
	}
	for _, opt := range options {
		opt(&opts)
	}

	report := Report{}
	err := json.Unmarshal(attestResult.Report, &report)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling report: %w", err)
	}

	signDataHash := sha256.Sum256([]byte(NoTEEMeasurement))
	if !ECDSAVerify(report.VerifyKey, signDataHash[:], report.Signature) {
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

	if opts.nonce != nil && !bytes.Equal(opts.nonce, report.Nonce) {
		return nil, fmt.Errorf("nonce mismatch: expected '%s', got '%s'",
			base64.StdEncoding.EncodeToString(opts.nonce),
			base64.StdEncoding.EncodeToString(report.Nonce),
		)
	}

	verifyResult := &VerifyResult{
		PublicKey: report.PublicKey,
		UserData: report.Userdata,
	}
	return verifyResult, nil
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
