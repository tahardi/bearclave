package tee_test

import (
	"context"
	crand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/tee"
)

func newTestAttester(t *testing.T) *tee.Attester {
	t.Helper()
	attester, err := tee.NewAttester(tee.NoTEE)
	require.NoError(t, err)
	return attester
}

func parseLeaf(t *testing.T, cert *tls.Certificate) *x509.Certificate {
	t.Helper()
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	require.NoError(t, err)
	return leaf
}

func verifyCertKeyBinding(t *testing.T, leaf *x509.Certificate) *tee.KeyBinding {
	t.Helper()
	verifier, err := tee.NewVerifier(tee.NoTEE)
	require.NoError(t, err)

	att, err := tee.ParseCertAttestation(leaf)
	require.NoError(t, err)

	binding, err := tee.VerifyKeyBinding(
		verifier,
		att,
		tee.TLSKeyBindingPurpose,
		tee.KeyBindingPolicy{Measurement: tee.NoTEEMeasurement, MaxAge: time.Hour},
	)
	require.NoError(t, err)
	return binding
}

func TestAttestedCertProvider_Interfaces(t *testing.T) {
	t.Run("CertProvider", func(_ *testing.T) {
		var _ tee.CertProvider = &tee.AttestedCertProvider{}
	})
}

func TestGenerateAttestedCert(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		attester := newTestAttester(t)
		privateKey := newTestECDSAPrivateKey(t)

		// when
		cert, err := tee.GenerateAttestedCert(
			attester,
			privateKey,
			tee.DefaultDomain,
			tee.DefaultIP,
			tee.DefaultAttestedCertValidity,
		)

		// then
		require.NoError(t, err)
		require.Equal(t, privateKey, cert.PrivateKey)

		leaf := parseLeaf(t, cert)
		binding := verifyCertKeyBinding(t, leaf)
		assert.Equal(t, leaf.RawSubjectPublicKeyInfo, binding.PublicKey)
		assert.Equal(t, leaf.NotBefore.Unix(), binding.IssuedAt)

		var found bool
		for _, ext := range leaf.Extensions {
			if ext.Id.Equal(tee.AttestedCertOID) {
				found = true
				assert.False(t, ext.Critical)
			}
		}
		assert.True(t, found)
	})
}

func TestAttestedCertProvider_GetCert(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		ctx := context.Background()
		provider, err := tee.NewAttestedCertProvider(
			newTestAttester(t),
			tee.DefaultDomain,
			tee.DefaultIP,
			tee.DefaultAttestedCertValidity,
		)
		require.NoError(t, err)

		// when
		cert, err := provider.GetCert(ctx)

		// then
		require.NoError(t, err)
		leaf := parseLeaf(t, cert)
		binding := verifyCertKeyBinding(t, leaf)
		assert.Equal(t, leaf.RawSubjectPublicKeyInfo, binding.PublicKey)
	})
}

func TestAttestedCertProvider_RotateCert(t *testing.T) {
	t.Run("happy path - new key and attestation", func(t *testing.T) {
		// given
		ctx := context.Background()
		provider, err := tee.NewAttestedCertProvider(
			newTestAttester(t),
			tee.DefaultDomain,
			tee.DefaultIP,
			tee.DefaultAttestedCertValidity,
		)
		require.NoError(t, err)

		oldCert, err := provider.GetCert(ctx)
		require.NoError(t, err)
		oldLeaf := parseLeaf(t, oldCert)

		// when
		err = provider.RotateCert(ctx)

		// then
		require.NoError(t, err)
		newCert, err := provider.GetCert(ctx)
		require.NoError(t, err)
		newLeaf := parseLeaf(t, newCert)

		assert.NotEqual(t, oldLeaf.RawSubjectPublicKeyInfo, newLeaf.RawSubjectPublicKeyInfo)
		binding := verifyCertKeyBinding(t, newLeaf)
		assert.Equal(t, newLeaf.RawSubjectPublicKeyInfo, binding.PublicKey)
	})
}

func TestParseCertAttestation(t *testing.T) {
	tests := []struct {
		name     string
		makeCert func(t *testing.T) *tls.Certificate
		wantErr  error
	}{
		{
			name: "happy path",
			makeCert: func(t *testing.T) *tls.Certificate {
				t.Helper()
				cert, err := tee.GenerateAttestedCert(
					newTestAttester(t),
					newTestECDSAPrivateKey(t),
					tee.DefaultDomain,
					tee.DefaultIP,
					tee.DefaultAttestedCertValidity,
				)
				require.NoError(t, err)
				return cert
			},
		},
		{
			name: "error - missing extension",
			makeCert: func(t *testing.T) *tls.Certificate {
				t.Helper()
				cert, err := tee.GenerateSelfSignedCert(
					newTestECDSAPrivateKey(t),
					tee.DefaultDomain,
					tee.DefaultIP,
					tee.DefaultValidity,
				)
				require.NoError(t, err)
				return cert
			},
			wantErr: tee.ErrCertProvider,
		},
		{
			name: "error - invalid extension json",
			makeCert: func(t *testing.T) *tls.Certificate {
				t.Helper()
				privateKey := newTestECDSAPrivateKey(t)
				template := x509.Certificate{
					SerialNumber: big.NewInt(1),
					Subject:      pkix.Name{CommonName: tee.DefaultDomain},
					NotBefore:    time.Now(),
					NotAfter:     time.Now().Add(time.Hour),
					ExtraExtensions: []pkix.Extension{
						{Id: tee.AttestedCertOID, Value: []byte("not json")},
					},
				}
				der, err := x509.CreateCertificate(
					crand.Reader,
					&template,
					&template,
					&privateKey.PublicKey,
					privateKey,
				)
				require.NoError(t, err)
				return &tls.Certificate{Certificate: [][]byte{der}, PrivateKey: privateKey}
			},
			wantErr: tee.ErrCertProvider,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			leaf := parseLeaf(t, tt.makeCert(t))

			// when
			got, err := tee.ParseCertAttestation(leaf)

			// then
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}
