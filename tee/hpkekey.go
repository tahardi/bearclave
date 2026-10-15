package tee

import (
	"crypto/hpke"
)

const (
	HPKEInfo                  = "bearclave/hpke/v1"
	HPKEResponseExportContext = "bearclave/hpke/v1/response"
	HPKEResponseSecretSize    = 32
)

type HPKEKey struct {
	privateKey hpke.PrivateKey
}

func NewHPKEKey() (*HPKEKey, error) {
	privateKey, err := hpke.MLKEM768X25519().GenerateKey()
	if err != nil {
		return nil, hpkeError("generating key", err)
	}
	return &HPKEKey{privateKey: privateKey}, nil
}

func (k *HPKEKey) PublicKey() []byte {
	return k.privateKey.PublicKey().Bytes()
}

func (k *HPKEKey) Open(msg SealedMessage, aad []byte) ([]byte, *HPKEResponseKey, error) {
	if len(msg.Enc) == 0 {
		return nil, nil, hpkeError("missing enc", nil)
	}

	recipient, err := hpke.NewRecipient(
		msg.Enc,
		k.privateKey,
		hpke.HKDFSHA256(),
		hpke.AES256GCM(),
		[]byte(HPKEInfo),
	)
	if err != nil {
		return nil, nil, hpkeError("making recipient", err)
	}

	plaintext, err := recipient.Open(aad, msg.Ciphertext)
	if err != nil {
		return nil, nil, hpkeError("opening request", err)
	}

	secret, err := recipient.Export(HPKEResponseExportContext, HPKEResponseSecretSize)
	if err != nil {
		return nil, nil, hpkeError("exporting response secret", err)
	}

	responseKey, err := NewHPKEResponseKey(secret)
	if err != nil {
		return nil, nil, err
	}
	return plaintext, responseKey, nil
}
