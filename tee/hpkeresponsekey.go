package tee

import (
	"crypto/aes"
	"crypto/cipher"
)

type HPKEResponseKey struct {
	aead cipher.AEAD
}

func NewHPKEResponseKey(secret []byte) (*HPKEResponseKey, error) {
	if len(secret) != HPKEResponseSecretSize {
		return nil, hpkeError("invalid response secret size", nil)
	}

	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, hpkeError("making cipher", err)
	}

	aead, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, hpkeError("making aead", err)
	}
	return &HPKEResponseKey{aead: aead}, nil
}

func (k *HPKEResponseKey) Seal(plaintext []byte) SealedMessage {
	return SealedMessage{
		Ciphertext: k.aead.Seal(nil, nil, plaintext, nil), //nolint:gosec // NewGCMWithRandomNonce generates the nonce
	}
}

func (k *HPKEResponseKey) Open(msg SealedMessage) ([]byte, error) {
	if len(msg.Enc) != 0 {
		return nil, hpkeError("unexpected enc in response", nil)
	}

	plaintext, err := k.aead.Open(nil, nil, msg.Ciphertext, nil)
	if err != nil {
		return nil, hpkeError("opening response", err)
	}
	return plaintext, nil
}
