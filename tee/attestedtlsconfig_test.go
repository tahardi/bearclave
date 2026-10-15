package tee_test

import (
	"context"
	"crypto/ecdsa"
	crand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/tee"
)

func newTestCert(
	t *testing.T,
	key *ecdsa.PrivateKey,
	notBefore time.Time,
	validity time.Duration,
	ext []pkix.Extension,
) *x509.Certificate {
	t.Helper()
	template := x509.Certificate{
		SerialNumber:    big.NewInt(notBefore.UnixNano()),
		Subject:         pkix.Name{CommonName: tee.DefaultDomain},
		NotBefore:       notBefore,
		NotAfter:        notBefore.Add(validity),
		ExtraExtensions: ext,
	}
	der, err := x509.CreateCertificate(crand.Reader, &template, &template, &key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return cert
}

func attestationExtension(
	t *testing.T,
	attester *tee.Attester,
	purpose string,
	spki []byte,
	issuedAt time.Time,
) pkix.Extension {
	t.Helper()
	att, err := tee.AttestKeyBinding(attester, tee.KeyBinding{
		Purpose:   purpose,
		PublicKey: spki,
		IssuedAt:  issuedAt.Unix(),
	})
	require.NoError(t, err)
	value, err := json.Marshal(att)
	require.NoError(t, err)
	return pkix.Extension{Id: tee.AttestedCertOID, Value: value}
}

func findAttestationExtension(t *testing.T, cert *x509.Certificate) pkix.Extension {
	t.Helper()
	for _, ext := range cert.Extensions {
		if ext.Id.Equal(tee.AttestedCertOID) {
			return ext
		}
	}
	require.FailNow(t, "missing attestation extension")
	return pkix.Extension{}
}

func newTestAttestedLeaf(t *testing.T, attester *tee.Attester, validity time.Duration) *x509.Certificate {
	t.Helper()
	cert, err := tee.GenerateAttestedCert(
		attester,
		newTestECDSAPrivateKey(t),
		tee.DefaultDomain,
		tee.DefaultIP,
		validity,
	)
	require.NoError(t, err)
	return parseLeaf(t, cert)
}

func TestVerifyAttestedCert(t *testing.T) {
	attester := newTestAttester(t)
	verifier, err := tee.NewVerifier(tee.NoTEE)
	require.NoError(t, err)
	nitroVerifier, err := tee.NewVerifier(tee.Nitro)
	require.NoError(t, err)

	policy := tee.KeyBindingPolicy{Measurement: tee.NoTEEMeasurement, MaxAge: time.Hour}
	withPolicy := func(modify func(*tee.KeyBindingPolicy)) tee.KeyBindingPolicy {
		p := policy
		modify(&p)
		return p
	}

	tests := []struct {
		name     string
		verifier *tee.Verifier
		makeCert func(t *testing.T) *x509.Certificate
		policy   tee.KeyBindingPolicy
		wantErr  error
	}{
		{
			name:     "happy path",
			verifier: verifier,
			makeCert: func(t *testing.T) *x509.Certificate {
				t.Helper()
				return newTestAttestedLeaf(t, attester, tee.DefaultAttestedCertValidity)
			},
			policy: policy,
		},
		{
			name:     "error - missing extension",
			verifier: verifier,
			makeCert: func(t *testing.T) *x509.Certificate {
				t.Helper()
				cert, err := tee.GenerateSelfSignedCert(
					newTestECDSAPrivateKey(t),
					tee.DefaultDomain,
					tee.DefaultIP,
					tee.DefaultValidity,
				)
				require.NoError(t, err)
				return parseLeaf(t, cert)
			},
			policy:  policy,
			wantErr: tee.ErrCertProvider,
		},
		{
			name:     "error - key binding mismatch",
			verifier: verifier,
			makeCert: func(t *testing.T) *x509.Certificate {
				t.Helper()
				certA := newTestAttestedLeaf(t, attester, tee.DefaultAttestedCertValidity)
				ext := findAttestationExtension(t, certA)
				return newTestCert(t, newTestECDSAPrivateKey(t), time.Now(), time.Hour, []pkix.Extension{ext})
			},
			policy:  policy,
			wantErr: tee.ErrKeyBindingPublicKey,
		},
		{
			name:     "error - tampered attestation",
			verifier: verifier,
			makeCert: func(t *testing.T) *x509.Certificate {
				t.Helper()
				key := newTestECDSAPrivateKey(t)
				spki, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
				require.NoError(t, err)
				ext := attestationExtension(t, attester, tee.TLSKeyBindingPurpose, spki, time.Now())

				var att tee.AttestResult
				require.NoError(t, json.Unmarshal(ext.Value, &att))
				var report map[string]any
				require.NoError(t, json.Unmarshal(att.Base.Report, &report))
				report["timestamp"] = time.Now().Add(-time.Minute).Unix()
				att.Base.Report, err = json.Marshal(report)
				require.NoError(t, err)
				ext.Value, err = json.Marshal(att)
				require.NoError(t, err)

				return newTestCert(t, key, time.Now(), time.Hour, []pkix.Extension{ext})
			},
			policy:  policy,
			wantErr: tee.ErrVerifier,
		},
		{
			name:     "error - wrong measurement",
			verifier: verifier,
			makeCert: func(t *testing.T) *x509.Certificate {
				t.Helper()
				return newTestAttestedLeaf(t, attester, tee.DefaultAttestedCertValidity)
			},
			policy:  withPolicy(func(p *tee.KeyBindingPolicy) { p.Measurement = "wrong" }),
			wantErr: tee.ErrVerifierMeasurement,
		},
		{
			name:     "error - wrong purpose",
			verifier: verifier,
			makeCert: func(t *testing.T) *x509.Certificate {
				t.Helper()
				key := newTestECDSAPrivateKey(t)
				spki, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
				require.NoError(t, err)
				ext := attestationExtension(t, attester, tee.HPKEKeyBindingPurpose, spki, time.Now())
				return newTestCert(t, key, time.Now(), time.Hour, []pkix.Extension{ext})
			},
			policy:  policy,
			wantErr: tee.ErrKeyBindingPurpose,
		},
		{
			name:     "error - stale binding",
			verifier: verifier,
			makeCert: func(t *testing.T) *x509.Certificate {
				t.Helper()
				return newTestAttestedLeaf(t, attester, 48*time.Hour)
			},
			policy:  withPolicy(func(p *tee.KeyBindingPolicy) { p.Now = time.Now().Add(25 * time.Hour) }),
			wantErr: tee.ErrKeyBindingStale,
		},
		{
			name:     "error - expired certificate",
			verifier: verifier,
			makeCert: func(t *testing.T) *x509.Certificate {
				t.Helper()
				return newTestAttestedLeaf(t, attester, tee.DefaultAttestedCertValidity)
			},
			policy: withPolicy(func(p *tee.KeyBindingPolicy) {
				p.Now = time.Now().Add(tee.DefaultAttestedCertValidity + time.Hour)
			}),
			wantErr: tee.ErrVerifier,
		},
		{
			name:     "error - wrong platform",
			verifier: nitroVerifier,
			makeCert: func(t *testing.T) *x509.Certificate {
				t.Helper()
				return newTestAttestedLeaf(t, attester, tee.DefaultAttestedCertValidity)
			},
			policy:  policy,
			wantErr: tee.ErrVerifier,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			cert := tt.makeCert(t)

			// when
			err := tee.VerifyAttestedCert(tt.verifier, cert, tt.policy)

			// then
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func newTestTLSServer(ctx context.Context, t *testing.T, provider tee.CertProvider) string {
	t.Helper()
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hello"))
	}))
	srv.TLS = &tls.Config{
		MinVersion: tls.VersionTLS13,
		GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			return provider.GetCert(ctx)
		},
	}
	srv.StartTLS()
	t.Cleanup(srv.Close)
	return strings.Replace(srv.URL, "127.0.0.1", tee.DefaultDomain, 1)
}

func getHello(ctx context.Context, t *testing.T, client *http.Client, url string) (string, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(body), nil
}

func TestNewAttestedTLSConfig(t *testing.T) {
	ctx := context.Background()
	verifier, err := tee.NewVerifier(tee.NoTEE)
	require.NoError(t, err)
	policy := tee.KeyBindingPolicy{Measurement: tee.NoTEEMeasurement, MaxAge: time.Hour}

	newAttestedProvider := func(t *testing.T) *tee.AttestedCertProvider {
		t.Helper()
		provider, err := tee.NewAttestedCertProvider(
			newTestAttester(t),
			tee.DefaultDomain,
			tee.DefaultIP,
			tee.DefaultAttestedCertValidity,
		)
		require.NoError(t, err)
		return provider
	}
	newClient := func(policy tee.KeyBindingPolicy) *http.Client {
		return &http.Client{Transport: &http.Transport{
			TLSClientConfig:   tee.NewAttestedTLSConfig(verifier, policy),
			DisableKeepAlives: true,
		}}
	}

	t.Run("happy path", func(t *testing.T) {
		// given
		url := newTestTLSServer(ctx, t, newAttestedProvider(t))
		client := newClient(policy)

		// when
		body, err := getHello(ctx, t, client, url)

		// then
		require.NoError(t, err)
		assert.Equal(t, "hello", body)
	})

	t.Run("happy path - after rotation", func(t *testing.T) {
		// given
		provider := newAttestedProvider(t)
		url := newTestTLSServer(ctx, t, provider)
		client := newClient(policy)
		_, err := getHello(ctx, t, client, url)
		require.NoError(t, err)
		require.NoError(t, provider.RotateCert(ctx))

		// when
		body, err := getHello(ctx, t, client, url)

		// then
		require.NoError(t, err)
		assert.Equal(t, "hello", body)
	})

	t.Run("error - plain self-signed server", func(t *testing.T) {
		// given
		provider, err := tee.NewSelfSignedCertProvider(tee.DefaultDomain, tee.DefaultIP, tee.DefaultValidity)
		require.NoError(t, err)
		url := newTestTLSServer(ctx, t, provider)
		client := newClient(policy)

		// when
		_, err = getHello(ctx, t, client, url)

		// then
		require.ErrorIs(t, err, tee.ErrCertProvider)
	})

	t.Run("error - wrong measurement", func(t *testing.T) {
		// given
		url := newTestTLSServer(ctx, t, newAttestedProvider(t))
		client := newClient(tee.KeyBindingPolicy{Measurement: "wrong", MaxAge: time.Hour})

		// when
		_, err := getHello(ctx, t, client, url)

		// then
		require.ErrorIs(t, err, tee.ErrVerifierMeasurement)
	})
}
