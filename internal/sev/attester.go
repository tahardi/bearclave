package sev

const AMD_SEV_USERDATA_SIZE = 64

type Attester struct{}

func NewAttester() (*Attester, error) {
	return &Attester{}, nil
}

func (n *Attester) Attest(userdata []byte) ([]byte, error) {
	return userdata, nil
}
