package tee

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"time"
)

func NewAttestedTLSConfig(verifier *Verifier, policy KeyBindingPolicy) *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS13,
		// The attestation check in VerifyConnection replaces CA chain verification.
		//nolint:gosec
		InsecureSkipVerify: true,
		VerifyConnection: func(cs tls.ConnectionState) error {
			if len(cs.PeerCertificates) == 0 {
				return verifierError("missing peer certificate", nil)
			}
			return VerifyAttestedCert(verifier, cs.PeerCertificates[0], policy)
		},
	}
}

func VerifyAttestedCert(verifier *Verifier, cert *x509.Certificate, policy KeyBindingPolicy) error {
	now := policy.Now
	if now.IsZero() {
		now = time.Now()
	}
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		return verifierError("certificate expired or not yet valid", nil)
	}

	att, err := ParseCertAttestation(cert)
	if err != nil {
		return err
	}

	binding, err := VerifyKeyBinding(verifier, att, TLSKeyBindingPurpose, policy)
	if err != nil {
		return err
	}

	if !bytes.Equal(binding.PublicKey, cert.RawSubjectPublicKeyInfo) {
		return keyBindingPublicKeyError("certificate key does not match attested key", nil)
	}
	return nil
}
