package notee_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tahardi/bearclave/internal/attestation/notee"
)

func TestAttester_Attest(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		privateKey := newTestPrivateKey(t)
		attester, err := notee.NewAttesterWithPrivateKey(privateKey)
		require.NoError(t, err)

		publicKey := append(
			privateKey.PublicKey.X.Bytes(),
			privateKey.PublicKey.Y.Bytes()...,
		)

		// ECDSA signatures are not deterministic, so we can't
		// test that the signature is equal to a specific value
		userdata := []byte("hello world")
		want := unsafe.Attestation{
			Userdata:  userdata,
			Publickey: publicKey,
		}

		// when
		attestationBytes, err := attester.Attest(userdata)
		require.NoError(t, err)

		got := unsafe.Attestation{}
		err = json.Unmarshal(attestationBytes, &got)

		// then
		require.NoError(t, err)
		assert.Equal(t, want.Publickey, got.Publickey)
		assert.Equal(t, want.Userdata, got.Userdata)
	})
}
