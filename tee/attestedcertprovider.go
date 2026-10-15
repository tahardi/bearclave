package tee

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/json"
	"sync"
	"time"
)

const DefaultAttestedCertValidity = 24 * time.Hour

var AttestedCertOID = asn1.ObjectIdentifier{1, 3, 9901, 1, 1}

type AttestedCertProvider struct {
	mu       sync.Mutex
	attester *Attester
	cert     *tls.Certificate
	domain   string
	ip       string
	validity time.Duration
}

func NewAttestedCertProvider(
	attester *Attester,
	domain string,
	ip string,
	validity time.Duration,
) (*AttestedCertProvider, error) {
	cert, err := generateAttestedCertWithNewKey(attester, domain, ip, validity)
	if err != nil {
		return nil, err
	}
	return &AttestedCertProvider{
		mu:       sync.Mutex{},
		attester: attester,
		cert:     cert,
		domain:   domain,
		ip:       ip,
		validity: validity,
	}, nil
}

func (p *AttestedCertProvider) GetCert(_ context.Context) (*tls.Certificate, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cert, nil
}

func (p *AttestedCertProvider) RotateCert(_ context.Context) error {
	cert, err := generateAttestedCertWithNewKey(p.attester, p.domain, p.ip, p.validity)
	if err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cert = cert
	return nil
}

func GenerateAttestedCert(
	attester *Attester,
	privateKey *ecdsa.PrivateKey,
	domain string,
	ip string,
	validity time.Duration,
) (*tls.Certificate, error) {
	template := makeCertTemplate(domain, ip, validity, time.Now())

	spki, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, certProviderError("marshaling public key", err)
	}

	att, err := AttestKeyBinding(attester, KeyBinding{
		Purpose:   TLSKeyBindingPurpose,
		PublicKey: spki,
		IssuedAt:  template.NotBefore.Unix(),
	})
	if err != nil {
		return nil, certProviderError("attesting key binding", err)
	}

	value, err := json.Marshal(att)
	if err != nil {
		return nil, certProviderError("marshaling attestation", err)
	}
	template.ExtraExtensions = []pkix.Extension{{Id: AttestedCertOID, Critical: false, Value: value}}

	der, err := x509.CreateCertificate(
		crand.Reader,
		&template,
		&template,
		&privateKey.PublicKey,
		privateKey,
	)
	if err != nil {
		return nil, certProviderError("creating certificate", err)
	}
	return &tls.Certificate{Certificate: [][]byte{der}, PrivateKey: privateKey}, nil
}

func ParseCertAttestation(cert *x509.Certificate) (*AttestResult, error) {
	for _, ext := range cert.Extensions {
		if !ext.Id.Equal(AttestedCertOID) {
			continue
		}
		var att AttestResult
		err := json.Unmarshal(ext.Value, &att)
		if err != nil {
			return nil, certProviderError("unmarshaling attestation", err)
		}
		return &att, nil
	}
	return nil, certProviderError("missing attestation extension", nil)
}

func generateAttestedCertWithNewKey(
	attester *Attester,
	domain string,
	ip string,
	validity time.Duration,
) (*tls.Certificate, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		return nil, certProviderError("generating private key", err)
	}
	return GenerateAttestedCert(attester, privateKey, domain, ip, validity)
}
