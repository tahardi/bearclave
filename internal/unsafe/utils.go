package unsafe

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"fmt"
	"math/big"
)

func ECDSASign(privateKey *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	r, s, err := ecdsa.Sign(crand.Reader, privateKey, data)
	if err != nil {
		return nil, fmt.Errorf("ecdsa signing data: %w", err)
	}
	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}

func ECDSAVerify(publicKey []byte, data []byte, signature []byte) bool {
	r := big.Int{}
	s := big.Int{}
	sigLen := len(signature)
	r.SetBytes(signature[:(sigLen / 2)])
	s.SetBytes(signature[(sigLen / 2):])

	x := big.Int{}
	y := big.Int{}
	keyLen := len(publicKey)
	x.SetBytes(publicKey[:(keyLen / 2)])
	y.SetBytes(publicKey[(keyLen / 2):])

	rawPubKey := ecdsa.PublicKey{Curve: elliptic.P256(), X: &x, Y: &y}
	return ecdsa.Verify(&rawPubKey, data, &r, &s)
}
