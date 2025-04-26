package nitro

import "fmt"

type Attester struct{}

func NewAttester() (*Attester, error) {
	return &Attester{}, nil
}

func (n *Attester) Attest(userdata []byte) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}
