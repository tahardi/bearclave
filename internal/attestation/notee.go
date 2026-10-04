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
	NoTeeMaxUserDataSize = 64
	NoTeeMeasurement     = "Not a TEE platform. Code measurements are not real."
	NoTeeValidityPeriod  = int64(31536000)

	bitsPerByte             = 8
	p256CoordinateSize      = 32
	p256CoordinateBits      = p256CoordinateSize * bitsPerByte
	uncompressedPointPrefix = 0x04
	prefixSize              = 1
	p256XOffset             = prefixSize
	p256YOffset             = p256XOffset + p256CoordinateSize
	p256UncompressedKeySize = p256YOffset + p256CoordinateSize
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
	Signature   *Signature `json:"signature"`
	VerifyKey   *PublicKey `json:"verifykey"`
	Timestamp   int64      `json:"timestamp"`
	Measurement string     `json:"measurement"`
}

func noTEESignedDigest(report *Report) ([]byte, error) {
	unsigned := *report
	unsigned.Signature = nil
	unsignedBytes, err := json.Marshal(unsigned)
	if err != nil {
		return nil, fmt.Errorf("marshaling unsigned report: %w", err)
	}
	digest := sha256.Sum256(unsignedBytes)
	return digest[:], nil
}

type NoTEEAttester struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *PublicKey
}

func NewNoTEEAttester() (*NoTEEAttester, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		return nil, attesterError("generating private key", err)
	}
	return NewNoTEEAttesterWithPrivateKey(privateKey)
}

func NewNoTEEAttesterWithPrivateKey(
	privateKey *ecdsa.PrivateKey,
) (*NoTEEAttester, error) {
	publicKey := &PublicKey{X: privateKey.X, Y: privateKey.Y}
	return &NoTEEAttester{privateKey: privateKey, publicKey: publicKey}, nil
}

func (a *NoTEEAttester) Close() error {
	return nil
}

func (a *NoTEEAttester) Attest(options ...AttestOption) (*AttestResult, error) {
	opts := MakeDefaultAttestOptions()
	for _, opt := range options {
		opt(&opts)
	}

	if len(opts.UserData) > NoTeeMaxUserDataSize {
		msg := fmt.Sprintf(
			"userdata must be less than %d bytes",
			NoTeeMaxUserDataSize,
		)
		return nil, attesterErrorUserData(msg, nil)
	}

	report := Report{
		Nonce:       opts.Nonce,
		Userdata:    opts.UserData,
		Signature:   nil,
		VerifyKey:   a.publicKey,
		Timestamp:   time.Now().Unix(),
		Measurement: NoTeeMeasurement,
	}
	digest, err := noTEESignedDigest(&report)
	if err != nil {
		return nil, attesterError("computing report digest", err)
	}

	signature, err := ECDSASign(a.privateKey, digest)
	if err != nil {
		return nil, attesterError("signing report", err)
	}
	report.Signature = signature

	reportBytes, err := json.Marshal(report)
	if err != nil {
		return nil, attesterError("marshaling report", err)
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
		Debug:       false,
		Measurement: "",
		Nonce:       nil,
		Timestamp:   time.Now(),
	}
	for _, opt := range options {
		opt(&opts)
	}

	report := Report{}
	err := json.Unmarshal(attestResult.Report, &report)
	if err != nil {
		return nil, verifierError("unmarshalling report", err)
	}

	digest, err := noTEESignedDigest(&report)
	if err != nil {
		return nil, verifierError("computing report digest", err)
	}

	err = ECDSAVerify(report.VerifyKey, digest, report.Signature)
	if err != nil {
		return nil, err
	}

	if opts.Timestamp.Unix() < report.Timestamp ||
		opts.Timestamp.Unix() > report.Timestamp+NoTeeValidityPeriod {
		return nil, verifierErrorTimestamp(
			"certificate has expired or is not yet valid",
			nil,
		)
	}

	if opts.Measurement != "" && opts.Measurement != report.Measurement {
		msg := fmt.Sprintf(
			"expected '%s' got '%s'",
			opts.Measurement,
			report.Measurement,
		)
		return nil, verifierErrorMeasurement(msg, nil)
	}

	if opts.Nonce != nil && !bytes.Equal(opts.Nonce, report.Nonce) {
		msg := fmt.Sprintf(
			"expected '%s' got '%s'",
			base64.StdEncoding.EncodeToString(opts.Nonce),
			base64.StdEncoding.EncodeToString(report.Nonce),
		)
		return nil, verifierErrorNonce(msg, nil)
	}

	verifyResult := &VerifyResult{
		UserData: report.Userdata,
	}
	return verifyResult, nil
}

func ECDSASign(privateKey *ecdsa.PrivateKey, data []byte) (*Signature, error) {
	r, s, err := ecdsa.Sign(crand.Reader, privateKey, data)
	if err != nil {
		return nil, verifierError("ecdsa signing data", err)
	}

	signature := &Signature{R: r, S: s}
	return signature, nil
}

func ECDSAVerify(publicKey *PublicKey, data []byte, signature *Signature) error {
	if publicKey == nil || publicKey.X == nil || publicKey.Y == nil {
		return verifierError("invalid public key", nil)
	}
	if signature == nil || signature.R == nil || signature.S == nil {
		return verifierError("invalid signature", nil)
	}

	coordinates := []*big.Int{publicKey.X, publicKey.Y}
	for _, c := range coordinates {
		if c.Sign() < 0 || c.BitLen() > p256CoordinateBits {
			return verifierError("invalid public key", nil)
		}
	}

	raw := make([]byte, p256UncompressedKeySize)
	raw[0] = uncompressedPointPrefix
	publicKey.X.FillBytes(raw[p256XOffset:p256YOffset])
	publicKey.Y.FillBytes(raw[p256YOffset:])
	key, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), raw)
	if err != nil {
		return verifierError("parsing public key", err)
	}

	if !ecdsa.Verify(key, data, signature.R, signature.S) {
		return verifierError("ecdsa verification failed", nil)
	}
	return nil
}
