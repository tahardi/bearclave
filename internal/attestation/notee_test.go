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
	t.Helper()
	attester, err := attestation.NewNoTEEAttester()
	require.NoError(t, err)

	report, err := attester.Attest(attestation.WithAttestUserData(userdata))
	require.NoError(t, err)

	return report, attestation.NoTeeMeasurement, time.Now()
}

func newTestPrivateKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)
	return privateKey
}

func TestNoTEE_Interfaces(t *testing.T) {
	t.Run("Attester", func(_ *testing.T) {
		var _ attestation.Attester = &attestation.NoTEEAttester{}
	})
	t.Run("Verifier", func(_ *testing.T) {
		var _ attestation.Verifier = &attestation.NoTEEVerifier{}
	})
}

func TestNoTEEAttester_Attest(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		privateKey := newTestPrivateKey(t)
		attester, err := attestation.NewNoTEEAttesterWithPrivateKey(privateKey)
		require.NoError(t, err)

		verifyKey := &attestation.PublicKey{X: privateKey.X, Y: privateKey.Y}

		// ECDSA signatures are not deterministic, so we can't
		// test that the signature is equal to a specific value
		userdata := []byte("hello world")
		want := attestation.Report{
			Userdata:  userdata,
			VerifyKey: verifyKey,
		}

		// when
		attestResult, err := attester.Attest(attestation.WithAttestUserData(userdata))
		require.NoError(t, err)

		got := attestation.Report{}
		err = json.Unmarshal(attestResult.Report, &got)

		// then
		require.NoError(t, err)
		assert.Equal(t, want.Userdata, got.Userdata)
	})

	t.Run("error - user data too long", func(t *testing.T) {
		// given
		privateKey := newTestPrivateKey(t)
		attester, err := attestation.NewNoTEEAttesterWithPrivateKey(privateKey)
		require.NoError(t, err)
		userdata := make([]byte, attestation.NoTeeMaxUserDataSize+1)

		// when
		_, err = attester.Attest(attestation.WithAttestUserData(userdata))

		// then
		require.ErrorIs(t, err, attestation.ErrAttesterUserData)
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
			attestation.WithVerifyMeasurement(measurement),
			attestation.WithVerifyTimestamp(timestamp),
		)

		// then
		require.NoError(t, err)
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
		require.NoError(t, err)
		assert.Equal(t, want, got.UserData)
	})

	t.Run("error - unmarshalling report", func(t *testing.T) {
		// given
		report := &attestation.AttestResult{Report: []byte("invalid report")}

		verifier, err := attestation.NewNoTEEVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifier)
		assert.ErrorContains(t, err, "unmarshalling report")
	})

	t.Run("error - verifying signature", func(t *testing.T) {
		report := attestation.Report{}
		reportBytes, err := json.Marshal(report)
		require.NoError(t, err)

		verifier, err := attestation.NewNoTEEVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(&attestation.AttestResult{Report: reportBytes})

		// then
		require.ErrorIs(t, err, attestation.ErrVerifier)
	})

	t.Run("error - expired report", func(t *testing.T) {
		// given
		want := []byte("hello world")
		timestamp := time.Unix(0, 0)
		report, _, _ := noTEEAttestation(t, want)

		verifier, err := attestation.NewNoTEEVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(report, attestation.WithVerifyTimestamp(timestamp))

		// then
		assert.ErrorIs(t, err, attestation.ErrVerifierTimestamp)
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
			attestation.WithVerifyMeasurement(wrongMeasurement),
			attestation.WithVerifyTimestamp(timestamp),
		)

		// then
		assert.ErrorIs(t, err, attestation.ErrVerifierMeasurement)
	})

	t.Run("error - wrong nonce", func(t *testing.T) {
		// given
		want := []byte("hello world")
		wrongNonce := []byte("wrong nonce")
		report, _, timestamp := noTEEAttestation(t, want)

		verifier, err := attestation.NewNoTEEVerifier()
		require.NoError(t, err)

		// when
		_, err = verifier.Verify(
			report,
			attestation.WithVerifyVerifyNonce(wrongNonce),
			attestation.WithVerifyTimestamp(timestamp),
		)

		// then
		assert.ErrorIs(t, err, attestation.ErrVerifierNonce)
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
		require.NoError(t, err)
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

		publicKey := &attestation.PublicKey{X: privateKey.X, Y: privateKey.Y}

		// when
		err = attestation.ECDSAVerify(publicKey, data, signature)

		// then
		require.NoError(t, err)
	})

	t.Run("error - invalid public key", func(t *testing.T) {
		// given
		data := []byte("Hello, World!")
		privateKey := newTestPrivateKey(t)
		signature, err := attestation.ECDSASign(privateKey, data)
		require.NoError(t, err)

		// when
		err = attestation.ECDSAVerify(nil, data, signature)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifier)
		assert.ErrorContains(t, err, "invalid public key")
	})

	t.Run("error - invalid signature", func(t *testing.T) {
		// given
		data := []byte("Hello, World!")
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
		require.NoError(t, err)

		publicKey := &attestation.PublicKey{X: privateKey.X, Y: privateKey.Y}

		// when
		err = attestation.ECDSAVerify(publicKey, data, nil)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifier)
		assert.ErrorContains(t, err, "invalid signature")
	})

	t.Run("error - wrong public key", func(t *testing.T) {
		// given
		data := []byte("Hello, World!")
		privateKey := newTestPrivateKey(t)
		signature, err := attestation.ECDSASign(privateKey, data)
		require.NoError(t, err)

		wrongPrivateKey := newTestPrivateKey(t)
		wrongPublicKey := &attestation.PublicKey{X: wrongPrivateKey.X, Y: wrongPrivateKey.Y}

		// when
		err = attestation.ECDSAVerify(wrongPublicKey, data, signature)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifier)
		assert.ErrorContains(t, err, "ecdsa verification failed")
	})

	t.Run("error - wrong data", func(t *testing.T) {
		// given
		data := []byte("Hello, World!")
		privateKey := newTestPrivateKey(t)
		signature, err := attestation.ECDSASign(privateKey, data)
		require.NoError(t, err)

		publicKey := &attestation.PublicKey{X: privateKey.X, Y: privateKey.Y}

		// when
		err = attestation.ECDSAVerify(publicKey, []byte("wrong data"), signature)

		// then
		require.ErrorIs(t, err, attestation.ErrVerifier)
		assert.ErrorContains(t, err, "ecdsa verification failed")
	})
}
