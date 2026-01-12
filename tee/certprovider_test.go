package tee_test

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/tee"
)

func newTestECDSAPrivateKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)
	return privateKey
}

func newTestEdDSAPrivateKey(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	_, privateKey, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)
	return privateKey
}

func newTestRSAPrivateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	privateKey, err := rsa.GenerateKey(crand.Reader, 3072)
	require.NoError(t, err)
	return privateKey
}

func TestSelfSignedCertProvider_Interfaces(t *testing.T) {
	t.Run("CertProvider", func(_ *testing.T) {
		var _ tee.CertProvider = &tee.SelfSignedCertProvider{}
	})
}

func TestNewSelfSignedCertProvider_GetCert(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		ctx := context.Background()
		privateKey := newTestECDSAPrivateKey(t)
		domain := tee.DefaultDomain
		ip := tee.DefaultIP
		validity := tee.DefaultValidity
		certProvider, err := tee.NewSelfSignedCertProviderWithKey(
			privateKey,
			domain,
			ip,
			validity,
		)
		require.NoError(t, err)

		// when
		cert, err := certProvider.GetCert(ctx)

		// then
		require.NoError(t, err)
		require.NotNil(t, cert)
	})
}

func TestNewSelfSignedCertProvider_RotateCert(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		ctx := context.Background()
		privateKey := newTestECDSAPrivateKey(t)
		domain := tee.DefaultDomain
		ip := tee.DefaultIP
		validity := tee.DefaultValidity
		certProvider, err := tee.NewSelfSignedCertProviderWithKey(
			privateKey,
			domain,
			ip,
			validity,
		)
		require.NoError(t, err)

		oldCert, err := certProvider.GetCert(ctx)
		require.NoError(t, err)

		// when
		err = certProvider.RotateCert(ctx)
		require.NoError(t, err)

		newCert, err := certProvider.GetCert(ctx)
		require.NoError(t, err)

		// then
		assert.NotEqual(t, oldCert, newCert)
	})
}

func TestGenerateSelfSignedCert(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		privateKey := newTestECDSAPrivateKey(t)
		domain := tee.DefaultDomain
		ip := tee.DefaultIP
		validity := tee.DefaultValidity

		// when
		cert, err := tee.GenerateSelfSignedCert(privateKey, domain, ip, validity)

		// then
		require.NoError(t, err)
		require.NotNil(t, cert)
		require.NotNil(t, cert.PrivateKey)
		require.Equal(t, privateKey, cert.PrivateKey)
		require.NotEmpty(t, cert.Certificate)

		x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
		require.NoError(t, err)

		// Verify certificate properties
		assert.Equal(t, domain, x509Cert.Subject.CommonName)
		assert.Contains(t, x509Cert.DNSNames, domain)
		assert.True(t, x509Cert.IPAddresses[0].Equal(net.ParseIP(ip)))
		assert.Equal(t, x509.KeyUsageKeyEncipherment|x509.KeyUsageDigitalSignature, x509Cert.KeyUsage)
		assert.Contains(t, x509Cert.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
		assert.True(t, x509Cert.BasicConstraintsValid)

		// Verify self-signed (issuer == subject)
		assert.Equal(t, x509Cert.Issuer, x509Cert.Subject)

		// Verify that cert can be used to verify itself
		certPool := x509.NewCertPool()
		certPool.AddCert(x509Cert)
		opts := x509.VerifyOptions{
			Roots:       certPool,
			DNSName:     domain,
			CurrentTime: time.Now(),
		}

		_, err = x509Cert.Verify(opts)
		assert.NoError(t, err)
	})
}

type unsupportedPrivateKey struct{}

func (u *unsupportedPrivateKey) Public() crypto.PublicKey {
	return nil
}

func TestGetPublicKey(t *testing.T) {
	t.Run("happy path - ecdsa", func(t *testing.T) {
		// given
		privateKey := newTestECDSAPrivateKey(t)
		want := privateKey.Public()

		// when
		got, err := tee.GetPublicKey(privateKey)

		// then
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("happy path - ed25519", func(t *testing.T) {
		// given
		privateKey := newTestEdDSAPrivateKey(t)
		want := privateKey.Public()

		// when
		got, err := tee.GetPublicKey(privateKey)

		// then
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("happy path - rsa", func(t *testing.T) {
		// given
		privateKey := newTestRSAPrivateKey(t)
		want := privateKey.Public()

		// when
		got, err := tee.GetPublicKey(privateKey)

		// then
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("error - unsupported key type", func(t *testing.T) {
		// given
		unsupportedKey := &unsupportedPrivateKey{}

		// when
		_, err := tee.GetPublicKey(unsupportedKey)

		// then
		require.ErrorIs(t, err, tee.ErrCertProvider)
		require.ErrorContains(t, err, "unsupported private key type")
	})
}
