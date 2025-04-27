package cvms

import "fmt"

type Verifier struct{}

func NewVerifier() (*Verifier, error) {
	return &Verifier{}, nil
}

func (n *Verifier) Verify(attestation []byte) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}
