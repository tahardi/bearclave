package attestation

type Attester interface {
	Attest(userdata []byte) (report []byte, err error)
}

