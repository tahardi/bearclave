package attestation

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"
)

const (
	NoTeeMeasurement    = "Not a TEE platform. Code measurements are not real."
	NoTeeValidityPeriod = int64(31536000)
)

var (
	ErrMeasurementMismatch = errors.New("measurement mismatch")
	ErrNonceMismatch       = errors.New("nonce mismatch")
	ErrInvalidCertificate  = errors.New("certificate has expired or is not yet valid")
	ErrInvalidPublicKey    = errors.New("invalid public key")
	ErrInvalidSignature    = errors.New("invalid signature")
	ErrECDSAVerification   = errors.New("ecdsa verification failed")
)

type PublicKey struct {
	X *big.Int `json:"x"`
	Y *big.Int `json:"y"`
}

type Signature struct {
	R *big.Int `json:"r"`
	S *big.Int `json:"s"`
}

type Report struct {
	Userdata    []byte     `json:"userdata"`
	Nonce       []byte     `json:"nonce"`
	PublicKey   []byte     `json:"publickey"`
	Signature   *Signature `json:"signature"`
	VerifyKey   *PublicKey `json:"verifykey"`
	Timestamp   int64      `json:"timestamp"`
	Measurement string     `json:"measurement"`
}

type NoTEEAttester struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *PublicKey
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
	publicKey := &PublicKey{X: privateKey.X, Y: privateKey.Y}
	return &NoTEEAttester{privateKey: privateKey, publicKey: publicKey}, nil
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

	signDataHash := sha256.Sum256([]byte(NoTeeMeasurement))
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
		Measurement: NoTeeMeasurement,
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
		debug:       false,
		measurement: "",
		nonce:       nil,
		timestamp:   time.Now(),
	}
	for _, opt := range options {
		opt(&opts)
	}

	report := Report{}
	err := json.Unmarshal(attestResult.Report, &report)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling report: %w", err)
	}

	signDataHash := sha256.Sum256([]byte(NoTeeMeasurement))
	err = ECDSAVerify(report.VerifyKey, signDataHash[:], report.Signature)
	if err != nil {
		return nil, fmt.Errorf("verifying signature: %w", err)
	}

	if opts.timestamp.Unix() < report.Timestamp ||
		opts.timestamp.Unix() > report.Timestamp+NoTeeValidityPeriod {
		return nil, ErrInvalidCertificate
	}

	if opts.measurement != "" && opts.measurement != report.Measurement {
		return nil, fmt.Errorf(
			"%w: expected '%s' got '%s'",
			ErrMeasurementMismatch,
			opts.measurement,
			report.Measurement,
		)
	}

	if opts.nonce != nil && !bytes.Equal(opts.nonce, report.Nonce) {
		return nil, fmt.Errorf("%w: expected '%s', got '%s'",
			ErrNonceMismatch,
			base64.StdEncoding.EncodeToString(opts.nonce),
			base64.StdEncoding.EncodeToString(report.Nonce),
		)
	}

	verifyResult := &VerifyResult{
		PublicKey: report.PublicKey,
		UserData:  report.Userdata,
	}
	return verifyResult, nil
}

func ECDSASign(privateKey *ecdsa.PrivateKey, data []byte) (*Signature, error) {
	r, s, err := ecdsa.Sign(crand.Reader, privateKey, data)
	if err != nil {
		return nil, fmt.Errorf("ecdsa signing data: %w", err)
	}

	signature := &Signature{R: r, S: s}
	return signature, nil
}

func ECDSAVerify(publicKey *PublicKey, data []byte, signature *Signature) error {
	if publicKey == nil || publicKey.X == nil || publicKey.Y == nil {
		return ErrInvalidPublicKey
	}
	if signature == nil || signature.R == nil || signature.S == nil {
		return ErrInvalidSignature
	}

	rawPubKey := ecdsa.PublicKey{Curve: elliptic.P256(), X: publicKey.X, Y: publicKey.Y}
	if !ecdsa.Verify(&rawPubKey, data, signature.R, signature.S) {
		return ErrECDSAVerification
	}
	return nil
}
