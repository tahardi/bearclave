package tee

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"
)

const (
	DefaultDomain = "localhost"
	DefaultIP     = "127.0.0.1"
	DefaultValidity = 365 * 24 * time.Hour
)

type CertProvider interface {
	GetCert(ctx context.Context) (*tls.Certificate, error)
	RotateCert(ctx context.Context) error
}

type SelfSignedCertProvider struct {
	mu         sync.Mutex
	cert       *tls.Certificate
	privateKey crypto.PrivateKey
	domain     string
	ip         string
	validity   time.Duration
}

func NewSelfSignedCertProvider(
	domain string,
	ip string,
	validity time.Duration,
) (*SelfSignedCertProvider, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		return nil, certProviderError("generating private key", err)
	}
	return NewSelfSignedCertProviderWithKey(privateKey, domain, ip, validity)
}

func NewSelfSignedCertProviderWithKey(
	privateKey crypto.PrivateKey,
	domain string,
	ip string,
	validity time.Duration,
) (*SelfSignedCertProvider, error) {
	cert, err := GenerateSelfSignedCert(privateKey, domain, ip, validity)
	if err != nil {
		return nil, err
	}
	return &SelfSignedCertProvider{
		mu:         sync.Mutex{},
		cert:       cert,
		privateKey: privateKey,
		domain:     domain,
		ip:         ip,
		validity:   validity,
	}, nil
}

func (s *SelfSignedCertProvider) GetCert(
	_ context.Context,
) (*tls.Certificate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cert, nil
}

func (s *SelfSignedCertProvider) RotateCert(
	_ context.Context,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cert, err := GenerateSelfSignedCert(
		s.privateKey,
		s.domain,
		s.ip,
		s.validity,
	)
	if err != nil {
		return err
	}
	s.cert = cert
	return nil
}

func GenerateSelfSignedCert(
	privateKey crypto.PrivateKey,
	domain string,
	ip string,
	validity time.Duration,
) (*tls.Certificate, error) {
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: domain,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(validity),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
		IPAddresses:           []net.IP{net.ParseIP(ip)},
	}

	publicKey, err := GetPublicKey(privateKey)
	if err != nil {
		return nil, err
	}

	certDER, err := x509.CreateCertificate(
		crand.Reader,
		&template,
		&template,
		publicKey,
		privateKey,
	)
	if err != nil {
		return nil, certProviderError("creating certificate", err)
	}

	return &tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  privateKey,
	}, nil
}

func GetPublicKey(privateKey crypto.PrivateKey) (crypto.PublicKey, error) {
	switch pk := privateKey.(type) {
	case *rsa.PrivateKey:
		return &pk.PublicKey, nil
	case *ecdsa.PrivateKey:
		return &pk.PublicKey, nil
	case ed25519.PrivateKey:
		return pk.Public(), nil
	default:
		msg := fmt.Sprintf("unsupported private key type: %T", privateKey)
		return nil, certProviderError(msg, nil)
	}
}
