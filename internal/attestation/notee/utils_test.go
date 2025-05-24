package notee_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/bearclave/internal/attestation/notee"
)

func newTestPrivateKey(t *testing.T) *ecdsa.PrivateKey {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)
	return privateKey
}

func TestECDSASign(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		data := []byte("Hello, World!")
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
		require.NoError(t, err)

		// when
		signature, err := notee.ECDSASign(privateKey, data)

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

		signature, err := notee.ECDSASign(privateKey, data)
		require.NoError(t, err)

		publicKey := append(
			privateKey.X.Bytes(),
			privateKey.Y.Bytes()...,
		)

		// when
		ok := notee.ECDSAVerify(publicKey, data, signature)

		// then
		assert.True(t, ok)
	})

	t.Run("wrong public key", func(t *testing.T) {
		// given
		data := []byte("Hello, World!")
		privateKey := newTestPrivateKey(t)
		signature, err := notee.ECDSASign(privateKey, data)
		require.NoError(t, err)

		wrongPrivateKey := newTestPrivateKey(t)
		wrongPublicKey := append(
			wrongPrivateKey.X.Bytes(),
			wrongPrivateKey.Y.Bytes()...,
		)

		// when
		ok := notee.ECDSAVerify(wrongPublicKey, data, signature)

		// then
		assert.False(t, ok)
	})

	t.Run("wrong data", func(t *testing.T) {
		// given
		data := []byte("Hello, World!")
		privateKey := newTestPrivateKey(t)
		signature, err := notee.ECDSASign(privateKey, data)
		require.NoError(t, err)

		publicKey := append(
			privateKey.X.Bytes(),
			privateKey.Y.Bytes()...,
		)

		// when
		ok := notee.ECDSAVerify(publicKey, []byte("wrong data"), signature)

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
		ok := notee.ECDSAVerify(publicKey, data, []byte("wrong signature"))

		// then
		assert.False(t, ok)
	})
}
