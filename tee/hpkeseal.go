package tee

import (
	"crypto/hpke"
)

func SealHPKERequest(
	publicKey []byte,
	plaintext []byte,
	aad []byte,
) (SealedMessage, *HPKEResponseKey, error) {
	pub, err := hpke.MLKEM768X25519().NewPublicKey(publicKey)
	if err != nil {
		return SealedMessage{}, nil, hpkeError("parsing public key", err)
	}

	enc, sender, err := hpke.NewSender(pub, hpke.HKDFSHA256(), hpke.AES256GCM(), []byte(HPKEInfo))
	if err != nil {
		return SealedMessage{}, nil, hpkeError("making sender", err)
	}

	ciphertext, err := sender.Seal(aad, plaintext)
	if err != nil {
		return SealedMessage{}, nil, hpkeError("sealing request", err)
	}

	secret, err := sender.Export(HPKEResponseExportContext, HPKEResponseSecretSize)
	if err != nil {
		return SealedMessage{}, nil, hpkeError("exporting response secret", err)
	}

	responseKey, err := NewHPKEResponseKey(secret)
	if err != nil {
		return SealedMessage{}, nil, err
	}
	return SealedMessage{Enc: enc, Ciphertext: ciphertext}, responseKey, nil
}
