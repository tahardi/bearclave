package attestation_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/internal/attestation"
)

func noTEEAttestation(
	t *testing.T,
	userdata []byte,
) (*attestation.AttestResult, string, time.Time) {
	attester, err := attestation.NewNoTEEAttester()
	require.NoError(t, err)

	report, err := attester.Attest(attestation.WithUserData(userdata))
	require.NoError(t, err)

	return report, attestation.NoTEEMeasurement, time.Now()
}

func newTestPrivateKey(t *testing.T) *ecdsa.PrivateKey {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)
	return privateKey
}

func TestNoTEE_Interfaces(t *testing.T) {
	t.Run("Attester", func(t *testing.T) {
		var _ attestation.Attester = &attestation.NoTEEAttester{}
	})
	t.Run("Verifier", func(t *testing.T) {
		var _ attestation.Verifier = &attestation.NoTEEVerifier{}
	})
}

func TestNoTEEAttester_Attest(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		privateKey := newTestPrivateKey(t)
		attester, err := attestation.NewNoTEEAttesterWithPrivateKey(privateKey)
		require.NoError(t, err)

		verifyKey := append(
			privateKey.X.Bytes(),
			privateKey.Y.Bytes()...,
		)

		// ECDSA signatures are not deterministic, so we can't
		// test that the signature is equal to a specific value
		userdata := []byte("hello world")
		want := attestation.Report{
			Userdata:  userdata,
			VerifyKey: verifyKey,
		}

		// when
		attestResult, err := attester.Attest(attestation.WithUserData(userdata))
		require.NoError(t, err)

		got := attestation.Report{}
		err = json.Unmarshal(attestResult.Report, &got)

		// then
		require.NoError(t, err)
		assert.Equal(t, want.Userdata, got.Userdata)
	})
}

func TestNoTEEVerifier_Verify(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		want := []byte("hello world")
		report, measurement, timestamp := noTEEAttestation(t, want)

		verifier, err := attestation.NewNoTEEVerifier()
		require.NoError(t, err)

		// when
		got, err := verifier.Verify(
			report,
			attestation.WithMeasurement(measurement),
			attestation.WithTimestamp(timestamp),
		)

		// then
		assert.NoError(t, err)
		assert.Equal(t, want, got.UserData)
	})

	t.Run("happy path - no measurement", func(t *testing.T) {
		// given
		want := []byte("hello world")
		report, _, _ := noTEEAttestation(t, want)

		verifier, err := attestation.NewNoTEEVerifier()
		require.NoError(t, err)

		// when
		got, err := verifier.Verify(report)

		// then
		assert.NoError(t, err)
		assert.Equal(t, want, got.UserData)
	})

	t.Run("error - invalid attestation report", func(t *testing.T) {
		// given
		report := &attestation.AttestResult{Report: []byte("invalid report")}

		verifier, err := attestation.NewNoTEEVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report)

		// then
		assert.ErrorContains(t, err, "unmarshaling report")
	})

	t.Run("error - invalid signature", func(t *testing.T) {
		// given
		report := attestation.Report{
			VerifyKey: []byte("verify key"),
			Signature: []byte("invalid signature"),
		}
		reportBytes, err := json.Marshal(report)
		require.NoError(t, err)

		verifier, err := attestation.NewNoTEEVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(&attestation.AttestResult{Report: reportBytes})

		// then
		assert.ErrorContains(t, err, "invalid signature")
	})

	t.Run("error - expired report", func(t *testing.T) {
		// given
		want := []byte("hello world")
		timestamp := time.Unix(0, 0)
		report, _, _ := noTEEAttestation(t, want)

		verifier, err := attestation.NewNoTEEVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report, attestation.WithTimestamp(timestamp))

		// then
		assert.ErrorContains(t, err, "certificate has expired or is not yet valid")
	})

	t.Run("error - wrong measurement", func(t *testing.T) {
		// given
		want := []byte("hello world")
		wrongMeasurement := "wrong measurement"
		report, _, timestamp := noTEEAttestation(t, want)

		verifier, err := attestation.NewNoTEEVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(
			report,
			attestation.WithMeasurement(wrongMeasurement),
			attestation.WithTimestamp(timestamp),
		)

		// then
		assert.ErrorContains(t, err, "measurement mismatch")
	})
}

func TestECDSASign(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		data := []byte("Hello, World!")
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
		require.NoError(t, err)

		// when
		signature, err := attestation.ECDSASign(privateKey, data)

		// then
		assert.NoError(t, err)
		assert.NotEmpty(t, signature)
	})
}

func TestECDSAVerify(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		data := []byte("Hello, World!")
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
		require.NoError(t, err)

		signature, err := attestation.ECDSASign(privateKey, data)
		require.NoError(t, err)

		publicKey := append(
			privateKey.X.Bytes(),
			privateKey.Y.Bytes()...,
		)

		// when
		ok := attestation.ECDSAVerify(publicKey, data, signature)

		// then
		assert.True(t, ok)
	})

	t.Run("wrong public key", func(t *testing.T) {
		// given
		data := []byte("Hello, World!")
		privateKey := newTestPrivateKey(t)
		signature, err := attestation.ECDSASign(privateKey, data)
		require.NoError(t, err)

		wrongPrivateKey := newTestPrivateKey(t)
		wrongPublicKey := append(
			wrongPrivateKey.X.Bytes(),
			wrongPrivateKey.Y.Bytes()...,
		)

		// when
		ok := attestation.ECDSAVerify(wrongPublicKey, data, signature)

		// then
		assert.False(t, ok)
	})

	t.Run("wrong data", func(t *testing.T) {
		// given
		data := []byte("Hello, World!")
		privateKey := newTestPrivateKey(t)
		signature, err := attestation.ECDSASign(privateKey, data)
		require.NoError(t, err)

		publicKey := append(
			privateKey.X.Bytes(),
			privateKey.Y.Bytes()...,
		)

		// when
		ok := attestation.ECDSAVerify(publicKey, []byte("wrong data"), signature)

		// then
		assert.False(t, ok)
	})

	t.Run("wrong signature", func(t *testing.T) {
		// given
		data := []byte("Hello, World!")
		privateKey := newTestPrivateKey(t)
		publicKey := append(
			privateKey.X.Bytes(),
			privateKey.Y.Bytes()...,
		)

		// when
		ok := attestation.ECDSAVerify(publicKey, data, []byte("wrong signature"))

		// then
		assert.False(t, ok)
	})
}
